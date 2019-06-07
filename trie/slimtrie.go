// Package trie provides SlimTrie implementation.
//
// A SlimTrie is a static, compressed Trie data structure.
// It removes unnecessary trie-node(single branch node etc).
// And it internally uses 3 compacted array to store a trie.
//
// SlimTrie memory overhead is about 14 bits per key(without value), or less.
//
// Key value map or key-range value map
//
// SlimTrie is natively something like a key value map.
// Actually besides as a key value map,
// to index a map of key range to value with SlimTrie is also very simple:
//
// Gives a set of key the same value, and use RangeGet() instead of Get().
// SlimTrie does not store branches for adjacent leaves with the same value.
//
// See SlimTrie.RangeGet .
//
// False Positive
//
// Just like Bloomfilter, SlimTrie does not contain full information of keys,
// thus there could be a false positive return:
// It returns some value and "true" but the key is not in there.
package trie

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"

	"github.com/blang/semver"
	"github.com/openacid/errors"
	"github.com/openacid/low/bitmap"
	"github.com/openacid/low/bitmap/varbits"
	"github.com/openacid/low/bitword"
	"github.com/openacid/low/pbcmpl"
	"github.com/openacid/low/sigbits"
	"github.com/openacid/low/tree"
	"github.com/openacid/low/vers"
	"github.com/openacid/slim/array"
	"github.com/openacid/slim/encode"
)

var (
	bw4 = bitword.BitWord[4]
)

const (
	// MaxNodeCnt is the max number of node. Node id in SlimTrie is int32.
	MaxNodeCnt = (1 << 31) - 1
)

// SlimTrie is a space efficient Trie index.
//
// The space overhead is about 14 bits per key and is irrelevant to key length.
//
// It does not store full key information, but only just enough info for
// locating a record.
// That's why an end user must re-validate the record key after reading it from
// other storage.
//
// It stores three parts of information in three SlimArray:
//
// `Children` stores node branches and children position.
// `Steps` stores the number of words to skip between a node and its parent.
// `Leaves` stores user data.
//
// Since 0.2.0
type SlimTrie struct {
	Children InnerNodes
	Steps    array.U16
	Leaves   array.Array
}

type versionedArray struct {
	*array.Base
}

func (va *InnerNodes) GetVersion() string {
	return slimtrieVersion
}

func (va *versionedArray) GetVersion() string {
	return slimtrieVersion
}

func (st *SlimTrie) GetVersion() string {
	return slimtrieVersion
}

func (st *SlimTrie) compatibleVersions() []string {
	return []string{
		"==1.0.0", // before 0.5.8 it is "1.0.0" for historical reason.
		"==0.5.8",
		"==0.5.9",
		"==" + slimtrieVersion,
	}
}

// NewSlimTrie create an SlimTrie.
// Argument e implements a encode.Encoder to convert user data to serialized
// bytes and back.
// Leave it nil if element in values are size fixed type and you do not really
// care about performance.
//	   int is not of fixed size.
//	   struct { X int64; Y int32; } hax fixed size.
//
// Since 0.2.0
func NewSlimTrie(e encode.Encoder, keys []string, values interface{}) (*SlimTrie, error) {
	return newSlimTrie(e, keys, values)
}

type subset struct {
	keyStart  int32
	keyEnd    int32
	fromIndex int32
}

func newSlimTrie(e encode.Encoder, keys []string, values interface{}) (*SlimTrie, error) {

	n := len(keys)
	if n == 0 {
		return emptySlimTrie(e), nil
	}

	for i := 0; i < len(keys)-1; i++ {
		if keys[i] >= keys[i+1] {
			return nil, errors.Wrapf(ErrKeyOutOfOrder,
				"keys[%d] >= keys[%d] %s %s", i, i+1, keys[i], keys[i+1])
		}
	}

	sb := sigbits.New(keys)
	// lencounts := countBitLen(keys)

	rvals := checkValues(reflect.ValueOf(values), n)
	tokeep := newValueToKeep(rvals)

	childi := make([]int32, 0, n)
	childv := make([][]int32, 0, n)
	childsize := make([]int32, 0, n)

	stepi := make([]int32, 0, n)
	stepv := make([]uint16, 0, n)

	leavesi := make([]int32, 0, n)
	leavesv := make([]interface{}, 0, n)

	queue := make([]subset, 0, n*2)
	queue = append(queue, subset{0, int32(n), 0})

	for i := 0; i < len(queue); i++ {
		nid := int32(i)
		o := queue[i]
		s, e := o.keyStart, o.keyEnd

		// single key, it is a leaf
		if e-s == 1 {
			if tokeep[s] {
				leavesi = append(leavesi, nid)
				leavesv = append(leavesv, getV(rvals, s))
			}
			continue
		}

		// need to create an inner node
		wordStart, distinctCnt := sb.Distinct(s, e, 16)

		wordsize := findPropWordSize(distinctCnt)

		bb := varbits.NewBitmapBuilder(wordsize)
		ks := make([]string, 0)
		for i := s; i < e; i++ {
			if tokeep[i] {
				ks = append(ks, keys[i])
			}
		}
		nodeBMIndexes := bb.BitPositions(ks, wordStart, true)

		// create inner node from following keys

		hasChildren := len(nodeBMIndexes) > 0

		if hasChildren {
			childi = append(childi, nid)
			childv = append(childv, nodeBMIndexes)
			childsize = append(childsize, varbits.BitmapSize(wordsize))

			// put keys with the same starting word to queue.

			for _, idx := range nodeBMIndexes {

				// Find the first key starting with label
				for ; s < e; s++ {
					kidx := bb.BitPos(keys[s], wordStart)
					if kidx == idx {
						break
					}
				}

				// Continue looking for the first key not starting with label
				var j int32
				for j = s + 1; j < e; j++ {
					kidx := bb.BitPos(keys[j], wordStart)
					if kidx != idx {
						break
					}
				}

				p := subset{
					keyStart:  s,
					keyEnd:    j,
					fromIndex: (wordStart + wordsize), // skip the label word
				}
				queue = append(queue, p)
				s = j
			}

			// Exclude the label word at parent node
			step := (wordStart - o.fromIndex)
			if step > 0xffff {
				panic(fmt.Sprintf("step=%d is too large. must < 2^16", step))
			}

			// By default to move 1 step forward, thus no need to store 1
			hasStep := step > 0
			if hasStep {
				stepi = append(stepi, nid)
				stepv = append(stepv, uint16(step))
			}
		}
	}

	nodeCnt := int32(len(queue))

	ch := newNodeArray(childi, childv, childsize)

	steps, err := array.NewU16(stepi, stepv)
	if err != nil {
		return nil, err
	}

	leaves := array.Array{}
	leaves.EltEncoder = e

	err = leaves.Init(leavesi, leavesv)
	if err != nil {
		return nil, errors.Wrapf(err, "failure init leaves")
	}

	// Avoid panic of slice index out of bound.
	ch.Index.Extend(nodeCnt)
	steps.ExtendIndex(nodeCnt)
	leaves.ExtendIndex(nodeCnt)

	st := &SlimTrie{
		Children: *ch,
		Steps:    *steps,
		Leaves:   leaves,
	}
	return st, nil
}

func checkValues(rvals reflect.Value, n int) reflect.Value {

	if rvals.Kind() != reflect.Slice {
		panic("values is not a slice")
	}

	valn := rvals.Len()

	if n != valn {
		panic(fmt.Sprintf("len(keys) != len(values): %d, %d", n, valn))
	}
	return rvals

}

// newValueToKeep creates a slice indicating which key to keep.
// Value of key[i+1] with the same value with key[i] do not need to keep.
func newValueToKeep(rvals reflect.Value) []bool {

	n := rvals.Len()

	tokeep := make([]bool, n)
	tokeep[0] = true

	for i := 0; i < n-1; i++ {
		tokeep[i+1] = getV(rvals, int32(i)+1) != getV(rvals, int32(i))
	}
	return tokeep
}

func getV(reflectSlice reflect.Value, i int32) interface{} {
	return reflectSlice.Index(int(i)).Interface()
}

func emptySlimTrie(e encode.Encoder) *SlimTrie {
	st := &SlimTrie{}
	st.Leaves.EltEncoder = e
	return st
}

// There are 2^x x-bits words, and 2^(x+1)-1 (0~x)-bits words.
// We need a 2^(x+1) bitmap to store all words with length from 0 to x
//
// Because upper nodes have more full x-bits words and lower nodes have more
// non-full x-bits words.
//
// Thus we separate full words and non-full words into 2 bitmap.
//
// For full words it requires 2^x bits
// For non-full words it requires 2^y - 1 bits, where y is the longest
// non-full words.
//
// Thus the entire bitmap size is 2^x + 2^y - 1, in binary form it is
// 100..00100..000
//
// Thus when we get the size we can extract 2 sizes from it by finding 2
// "1" in it.
//
// And we use 1 bit to indicate which one is the full-word bitmap.
//
// Encoding a binary tree
// In a tree, leaf nodes are x-bits word, inner nodes are words with less
// than x.
//
// Recursively put nodes in order:
// node, left-tree, right-right
//
// to locate a bit for a searching path p:
// index = 2 p + rank0(p)
//
// p may have less than x steps for a inner node.
func findPropWordSize(distinctCnt []int32) int32 {

	l := int32(len(distinctCnt))
	max := distinctCnt[l-1]

	// Ratio of storing words directly to storing with bitmap, for every word size in percentage:
	// Ratio = actual-data-size / bitmap-size
	// If storing them requires less space than with a bitmap, stop looking
	// up.
	// If not a lot space is wasted, we prefer to use as a larger bitmap as
	// possible.
	ratio := make([]int32, l)

	for i := int32(1); i < l; i++ {
		width := i
		bmsize := varbits.BitmapSize(width)
		ratio[i] = distinctCnt[i] * width * 100 / bmsize
	}

	rst := int32(1)
	for ; rst < l && distinctCnt[rst] == 1; rst++ {
	}
	for ; rst < l-1 && ratio[rst+1] >= 25 && distinctCnt[rst] < max; rst++ {
	}
	for ; rst > 1 && distinctCnt[rst] == distinctCnt[rst-1]; rst-- {

	}
	return rst
}

// rst[i] means the number of keys of length i, in bit.
func countBitLen(keys []string) []int32 {
	max := 0
	for _, k := range keys {
		l := len(k) * 8

		if max < l {
			max = l
		}
	}
	rst := make([]int32, max+1)
	for _, k := range keys {
		l := len(k) * 8
		rst[l]++
	}
	return rst
}

// RangeGet look for a range that contains a key in SlimTrie.
//
// A range that contains a key means range-start <= key <= range-end.
//
// It returns the value the range maps to, and a bool indicate if a range is
// found.
//
// A positive return value does not mean the range absolutely exists, which in
// this case, is a "false positive".
//
// Since 0.4.3
func (st *SlimTrie) RangeGet(key string) (interface{}, bool) {

	lID, eqID, _ := st.searchID(key)

	// an "equal" macth means key is a prefix of either start or end of a range.
	if eqID != -1 {
		v, found := st.Leaves.Get(eqID)
		if found {
			return v, found
		}

		// else: maybe matched at a inner node.
	}

	// key is smaller than any range-start or range-end.
	if lID == -1 {
		return nil, false
	}

	// Preceding value is the start of this range.
	// It might be a false-positive

	lVal, _ := st.Leaves.Get(lID)
	return lVal, true
}

// Search for a key in SlimTrie.
//
// It returns values of 3 values:
// The value of greatest key < `key`. It is nil if `key` is the smallest.
// The value of `key`. It is nil if there is not a matching.
// The value of smallest key > `key`. It is nil if `key` is the greatest.
//
// A non-nil return value does not mean the `key` exists.
// An in-existent `key` also could matches partial info stored in SlimTrie.
//
// Since 0.2.0
func (st *SlimTrie) Search(key string) (lVal, eqVal, rVal interface{}) {

	lID, eqID, rID := st.searchID(key)

	if lID != -1 {
		lVal, _ = st.Leaves.Get(lID)
	}
	if eqID != -1 {
		eqVal, _ = st.Leaves.Get(eqID)
	}
	if rID != -1 {
		rVal, _ = st.Leaves.Get(rID)
	}

	return
}

// searchID searches for key and returns 3 leaf node id:
//
// The id of greatest key < `key`. It is -1 if `key` is the smallest.
// The id of `key`. It is -1 if there is not a matching.
// The id of smallest key > `key`. It is -1 if `key` is the greatest.
func (st *SlimTrie) searchID(key string) (lID, eqID, rID int32) {
	lID, eqID, rID = -1, 0, -1

	lenWords := int32(8 * len(key))

	for i := int32(0); i <= lenWords; {

		hasChildren := st.Children.Index.Has(eqID)
		if !hasChildren {
			break
		}

		step, _ := st.Steps.Get(eqID)
		i += int32(step)
		if i > lenWords {
			rID = eqID
			eqID = -1
			break
		}

		var l, r int32
		var bmsize int32
		var childID int32

		from, to := st.Children.nodeBitRange(eqID)

		bmsize, l, childID = st.Children.getLabelRank(eqID, key, i)

		// TODO l, r check boundary
		r = childID + 1

		lparent := st.Children.getParentBitPos(l)
		rparent := st.Children.getParentBitPos(r)

		if lparent >= from && lparent < to {
			lID = l
		}
		if rparent >= from && rparent < to {
			rID = r
		}

		if l == childID {
			eqID = -1
			break
		}
		eqID = childID
		i += varbits.MaxbitsOfBitmap(bmsize)
	}

	if eqID != -1 {

		hasChildren := st.Children.Index.Has(eqID)
		if hasChildren {
			rID = eqID
			eqID = -1
		}
	}

	if lID != -1 {
		lID = st.rightMost(lID)
	}
	if rID != -1 {
		rID = st.leftMost(rID)
	}

	return
}

// just return equal value for trie.Search benchmark

// Get the value of the specified key from SlimTrie.
//
// If the key exist in SlimTrie, it returns the correct value.
// If the key does NOT exist in SlimTrie, it could also return some value.
//
// Because SlimTrie is a "index" but not a "kv-map", it does not stores complete
// info of all keys.
// SlimTrie tell you "WHERE IT POSSIBLY BE", rather than "IT IS JUST THERE".
//
// Since 0.2.0
func (st *SlimTrie) Get(key string) (eqVal interface{}, found bool) {

	found = false
	eqID := int32(0)

	// count in bit
	lenWords := int32(8 * len(key))

	for idx := int32(0); idx <= lenWords; {

		hasInner := st.Children.Index.Has(eqID)
		if !hasInner {
			// maybe a leaf
			break
		}

		step, _ := st.Steps.Get(eqID)
		idx += int32(step)

		var xx int32
		var bmsize int32
		bmsize, xx, eqID = st.Children.getLabelRank(eqID, key, idx)
		if xx == eqID {
			// no such branch of label
			return nil, false
		}
		idx += varbits.MaxbitsOfBitmap(bmsize)
	}

	if eqID != -1 {
		eqVal, found = st.Leaves.Get(eqID)
	}

	return
}

func (st *SlimTrie) getChild(idx int32) (bitmap uint16, offset int32, found bool) {
	bm, rank, found := st.Children.GetWithRank(idx)
	if found {
		return uint16(bm), rank + 1, true
	}
	return 0, 0, false
}

func (st *SlimTrie) getStep(idx int32) uint16 {
	step, _ := st.Steps.Get(idx)
	// 1 for the label word at parent node.
	return step + 1
}

func (st *SlimTrie) leftMost(idx int32) int32 {
	for {

		if st.Leaves.Has(idx) {
			return idx
		}

		b := st.Children
		from, _ := b.nodeBitRange(idx)

		r0, _ := bitmap.Rank128(b.Elts.Words, b.Elts.RankIndex, from)
		idx = r0 + 1
	}
}

func (st *SlimTrie) rightMost(idx int32) int32 {
	for {
		if st.Leaves.Has(idx) {
			return idx
		}

		b := st.Children
		_, to := b.nodeBitRange(idx)
		_, idx = bitmap.Rank128(b.Elts.Words, b.Elts.RankIndex, to-1)
	}
}

// Marshal serializes it to byte stream.
//
// Since 0.4.3
func (st *SlimTrie) Marshal() ([]byte, error) {
	var buf []byte
	writer := bytes.NewBuffer(buf)

	_, err := pbcmpl.Marshal(writer, &st.Children)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to marshal children")
	}

	_, err = pbcmpl.Marshal(writer, &versionedArray{&st.Steps.Base})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to marshal steps")
	}

	_, err = pbcmpl.Marshal(writer, &versionedArray{&st.Leaves.Base})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to marshal leaves")
	}

	return writer.Bytes(), nil
}

// Unmarshal de-serializes and ratio SlimTrie from a byte stream.
//
// Since 0.4.3
func (st *SlimTrie) Unmarshal(buf []byte) error {

	var ver string
	var err error
	var children *array.Bitmap16
	compatible := st.compatibleVersions()
	reader := bytes.NewReader(buf)

	_, h, err := pbcmpl.ReadHeader(reader)
	if err != nil {
		return errors.WithMessage(err, "failed to unmarshal header")
	}

	ver = h.GetVersion()

	if !vers.IsCompatible(ver, compatible) {
		return errors.Wrapf(ErrIncompatible,
			fmt.Sprintf(`version: "%s", compatible versions:"%s"`,
				ver,
				strings.Join(compatible, " || ")))
	}

	reader = bytes.NewReader(buf)

	if checkVer(ver, "==1.0.0 || <0.5.10") {
		children = &array.Bitmap16{}
		_, _, err = pbcmpl.Unmarshal(reader, children)
	} else {
		_, _, err = pbcmpl.Unmarshal(reader, &st.Children)
	}
	if err != nil {
		return errors.WithMessage(err, "failed to unmarshal children")
	}

	_, _, err = pbcmpl.Unmarshal(reader, &st.Steps)
	if err != nil {
		return errors.WithMessage(err, "failed to unmarshal steps")
	}

	_, _, err = pbcmpl.Unmarshal(reader, &st.Leaves)
	if err != nil {
		return errors.WithMessage(err, "failed to unmarshal leaves")
	}

	// backward compatible:

	before058(st, ver, children)
	eq058(st, ver, children)
	eq059(st, ver, children)

	return nil
}

func checkVer(ver string, want string) bool {
	v, err := semver.Parse(ver)
	if err != nil {
		panic(err)
	}

	chk, err := semver.ParseRange(want)
	if err != nil {
		panic(err)
	}

	return chk(v)
}

func before058(st *SlimTrie, ver string, ch *array.Bitmap16) {
	if !checkVer(ver, "==1.0.0 || <0.5.8") {
		return
	}
	before000510Step(st, ver, ch)
	before000510ToNewChildrenArray(st, ver, ch)
	before000509ExtendBitmapIndex(st, ver, ch)
}

func eq058(st *SlimTrie, ver string, ch *array.Bitmap16) {
	if !checkVer(ver, "==0.5.8") {
		return
	}
	before000510Step(st, ver, ch)
	before000510ToNewChildrenArray(st, ver, ch)
	before000509ExtendBitmapIndex(st, ver, ch)
}

func eq059(st *SlimTrie, ver string, ch *array.Bitmap16) {
	if !checkVer(ver, "==0.5.9") {
		return
	}
	before000510Step(st, ver, ch)
	before000510ToNewChildrenArray(st, ver, ch)
}

func before000510ToNewChildrenArray(st *SlimTrie, ver string, ch *array.Bitmap16) {

	// 1.0.0 is the initial version.
	// From 0.5.8 it starts writing version to marshaled data.
	// In 0.5.4 it starts using Bitmap to store Children elements.
	// But 0.5.4 marshals data with version == 1.0.0

	// Convert u32 node to ranked bitmap node

	if checkVer(ver, "==1.0.0 || <0.5.10") {

		// There are two format with version 1.0.0:
		// Before 0.5.4 Children elements are in Elts
		// From 0.5.4 Children elements are in BMElts

		//  rebuild nodes

		childi := make([]int32, 0)
		childv := make([][]int32, 0)
		childsize := make([]int32, 0)

		stepi := make([]int32, 0)
		stepv := make([]uint16, 0)

		leavesi := make([]int32, 0)
		leavesv := make([]interface{}, 0)

		type qq struct {
			oldid  int32
			isLeaf bool
		}

		q := make([]*qq, 0)
		q = append(q, &qq{oldid: 0, isLeaf: false})

		nextAddOldid := int32(1)

		for newid := int32(0); newid < int32(len(q)); newid++ {
			qelt := q[newid]

			if qelt.isLeaf {
				continue
			}

			oldid := qelt.oldid
			hasLeaf := st.Leaves.Has(oldid)
			hasInner := ch.Has(oldid)

			if hasInner {
				if hasLeaf {
					// an inner node is also play as a leaf. In 0.5.10, separate them.
					q = append(q, &qq{oldid: oldid, isLeaf: true})
				}
				words := getOldChild(ch, oldid)
				for range words {
					ee := &qq{oldid: nextAddOldid, isLeaf: false}
					q = append(q, ee)
					nextAddOldid++
				}
			} else {
				if hasLeaf {
					qelt.isLeaf = true
				} else {
					panic("no inner no leaf??")
				}
			}
		}

		for newid, qelt := range q {
			oldid := qelt.oldid
			if !qelt.isLeaf {
				childi = append(childi, int32(newid))
				words := getOldChild(ch, oldid)
				vbm := toVarBMIndexes(words)
				if st.Leaves.Has(oldid) {
					// "" is a explicit branch in 0.5.10
					vbm = append([]int32{0}, vbm...)
				}
				childv = append(childv, vbm)
				childsize = append(childsize, 32)

			}

			if !qelt.isLeaf && st.Steps.Has(oldid) {
				stepi = append(stepi, int32(newid))
				stp, found := st.Steps.Get(oldid)
				if !found {
					panic("step not found??")
				}
				stepv = append(stepv, stp)
			}

			if qelt.isLeaf && st.Leaves.Has(oldid) {
				leavesi = append(leavesi, int32(newid))
				lv, found := st.Leaves.Get(oldid)
				if !found {
					panic("leaf not found??")
				}
				leavesv = append(leavesv, lv)
			}

		}

		// create
		{

			ch := newNodeArray(childi, childv, childsize)

			steps, err := array.NewU16(stepi, stepv)
			if err != nil {
				panic("create steps error??")
			}

			leaves := array.Array{}
			leaves.EltEncoder = st.Leaves.EltEncoder

			err = leaves.Init(leavesi, leavesv)
			if err != nil {
				panic("create leaves error??")
			}
			st.Children = *ch
			st.Steps = *steps
			st.Leaves = leaves

		}

	}
}

func toVarBMIndexes(words []int32) []int32 {
	varbm := make([]int32, len(words))
	for i, pos := range words {
		b := varbits.New(uint64(pos), 4, 4)
		varbm[i] = varbits.ToIndex(b)
	}

	return varbm
}

func getOldChild(ch *array.Bitmap16, idx int32) []int32 {

	endian := binary.LittleEndian

	var barray []int32
	eltIdx, found := ch.GetEltIndex(idx)

	if !found {
		panic("not found index???")
	}

	if ch.Flags&array.ArrayFlagIsBitmap == 0 {

		// load from Base.Elts

		v := endian.Uint32(ch.Elts[eltIdx*4:])
		barray = bitmap.ToArray([]uint64{uint64(v & 0xffff)})

	} else {

		// load from Base.BMElts

		wordI := eltIdx * 16 / 64
		j := eltIdx * 16 % 64
		v := ch.BMElts.Words[wordI] >> uint(j)
		barray = bitmap.ToArray([]uint64{v & 0xffff})
	}
	return barray
}

func before000509ExtendBitmapIndex(st *SlimTrie, ver string, ch *array.Bitmap16) {

	// From 0.5.9 it create aligned array bitmap.
	// Need to align array bitmap for previous versions.

	if checkVer(ver, "==1.0.0 || <0.5.9") {

		n := 0
		if n < len(ch.Bitmaps)*64 {
			n = len(ch.Bitmaps) * 64
		}
		if n < len(st.Steps.Bitmaps)*64 {
			n = len(st.Steps.Bitmaps) * 64
		}
		if n < len(st.Leaves.Bitmaps)*64 {
			n = len(st.Leaves.Bitmaps) * 64
		}

		nodeCnt := int32(n)
		st.Children.Index.Extend(nodeCnt)
		st.Steps.ExtendIndex(nodeCnt)
		st.Leaves.ExtendIndex(nodeCnt)
	}
}

func before000510Step(st *SlimTrie, ver string, ch *array.Bitmap16) {

	// From 0.5.10 step does not include the count of the label word.
	// From 0.5.10 step is in bit instead of in 4-bit.

	if checkVer(ver, "==1.0.0 || <0.5.10") {

		ii := st.Steps.Indexes()
		for i, idx := range ii {
			v, found := st.Steps.Get(idx)
			if !found {
				panic("not found??")
			}

			v -= 1

			if v > 0xffff/4 {
				panic("old step is too large")
			}

			v *= 4

			st.Steps.Elts[i*2] = byte(v)
			st.Steps.Elts[i*2+1] = byte(v >> 8)
		}
	}
}

// Reset implements proto.Message
//
// Since 0.4.3
func (st *SlimTrie) Reset() {
	st.Children.Reset()
	st.Steps.Array32.Reset()
	st.Leaves.Array32.Reset()
}

// String implements proto.Message and output human readable multiline
// representation.
//
// A node is in form of
//   <income-label>-><node-id>+<step>*<fanout-count>=<value>
// E.g.:
//   000->#000+2*3
//            001->#001+4*2
//                     003->#004+1=0
//                              006->#007+2=1
//                     004->#005+1=2
//                              006->#008+2=3
//            002->#002+3=4
//                     006->#006+2=5
//                              006->#009+2=6
//            003->#003+5=7`[1:]
//
// Since 0.4.3
func (st *SlimTrie) String() string {
	s := &slimTrieStringly{
		st: st,
		inners: st.Children.Index.
			ToArray(),
		labels: make(map[int32]map[string]int32),
	}
	ch := st.Children
	for _, n := range s.inners {
		labels := ch.labels(n)

		s.labels[n] = make(map[string]int32)
		for _, l := range labels {
			idx := varbits.ToIndex(l)
			_, r1 := s.st.Children.getChildRank(n, idx)
			lstr := varbits.Str(l)
			s.labels[n][lstr] = r1
		}
	}

	return tree.String(s)
}

// ProtoMessage implements proto.Message
//
// Since 0.4.3
func (st *SlimTrie) ProtoMessage() {}
