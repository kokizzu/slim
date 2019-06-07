package trie

import (
	"github.com/openacid/low/bitmap"
	"github.com/openacid/low/bitmap/varbits"
)

func newNodeArray(index []int32, elts [][]int32, eltWidth []int32) *InnerNodes {

	b := &InnerNodes{}

	// if an elt exist
	exbm := bitmap.Of(index)
	b.Index = &Bitmap{
		Words:     exbm,
		RankIndex: bitmap.IndexRank64(exbm),
	}

	// where an elt is and the size
	ps := stepToPos(eltWidth)
	idxbm := bitmap.Of(ps)
	b.Starts = &Bitmap{
		Words:       idxbm,
		SelectIndex: bitmap.IndexSelect64(idxbm),
	}

	// elts
	eltbm := bitmap.Join(elts, eltWidth)
	b.Elts = &Bitmap{
		Words:       eltbm,
		RankIndex:   bitmap.IndexRank128(eltbm),
		SelectIndex: bitmap.IndexSelect64(eltbm),
	}

	return b
}

func stepToPos(steps []int32) []int32 {

	n := int32(len(steps))
	ps := make([]int32, n+1)
	p := int32(0)
	for i := int32(0); i < n; i++ {
		ps[i] = p
		// align to 4 bit
		p += steps[i] / 4
	}
	ps[n] = p
	return ps
}

func (b *InnerNodes) GetWithRank(idx int32) (uint64, int32, bool) {

	ithElt, iplus1Elt := bitmap.Rank64(b.Index.Words, b.Index.RankIndex, idx)
	if ithElt == iplus1Elt {
		return 0, 0, false
	}

	eltBitIdx := bitmap.Select(b.Starts.Words, b.Starts.SelectIndex, ithElt)
	eltBitIdx2 := bitmap.Select(b.Starts.Words, b.Starts.SelectIndex, iplus1Elt)
	eltLen := eltBitIdx2 - eltBitIdx

	iWord := eltBitIdx >> 6
	wordI2 := eltBitIdx2 >> 6

	j := eltBitIdx & 63
	w := b.Elts.Words[iWord]

	v := (w >> uint(j))
	if wordI2 > iWord {
		w2 := b.Elts.Words[wordI2]
		v &= (1 << uint(64-j)) - 1
		v |= w2 << uint(64-j)
	}

	v &= (1 << uint(eltLen)) - 1

	_, rank := bitmap.Rank128(b.Elts.Words, b.Elts.RankIndex, eltBitIdx)

	return v, rank, true
}

func (b *InnerNodes) nodeBitRange(idx int32) (int32, int32) {

	ithElt, iplus1Elt := bitmap.Rank64(b.Index.Words, b.Index.RankIndex, idx)
	if ithElt == iplus1Elt {
		panic("no such elt")
	}

	from := bitmap.Select(b.Starts.Words, b.Starts.SelectIndex, ithElt)
	to := bitmap.Select(b.Starts.Words, b.Starts.SelectIndex, iplus1Elt)

	return from * 4, to * 4
}

func (b *InnerNodes) labels(idx int32) []uint64 {
	from, to := b.nodeBitRange(idx)
	bmsize := to - from
	bm := bitmap.Slice(b.Elts.Words, from, to)

	vbs := varbits.Parse(bmsize, bm)
	return vbs
}

func (b *InnerNodes) getChildRank(idx int32, label int32) (int32, int32) {

	from, _ := b.nodeBitRange(idx)

	return bitmap.Rank128(b.Elts.Words, b.Elts.RankIndex, from+label)
}

func (b *InnerNodes) getLabelRank(idx int32, key string, ki int32) (int32, int32, int32) {

	from, to := b.nodeBitRange(idx)

	// elt length
	l := to - from

	wordsize := varbits.MaxbitsOfBitmap(l)
	kidx := varbits.ToIndex(varbits.FromStr(key, ki, wordsize))

	r0, r1 := bitmap.Rank128(b.Elts.Words, b.Elts.RankIndex, from+kidx)
	return to - from, r0, r1
}

func (b *InnerNodes) getParentBitPos(idx int32) int32 {
	return bitmap.Select(b.Elts.Words, b.Elts.SelectIndex, idx-1)
}
