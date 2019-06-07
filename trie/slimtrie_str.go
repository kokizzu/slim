package trie

import (
	"fmt"
	"sort"
)

// slimTrieStringly is a wrapper that implements tree.Tree .
// It is a helper to convert SlimTrie to string.
//
// Since 0.5.1
type slimTrieStringly struct {
	st     *SlimTrie
	inners []int32
	// node, label(varbits), node
	labels map[int32]map[string]int32
}

// Child implements tree.Tree
//
// Since 0.5.1
func (s *slimTrieStringly) Child(node, branch interface{}) interface{} {

	n := stNodeID(node)
	b := branch.(string)
	return s.labels[n][b]
}

// Labels implements tree.Tree
//
// Since 0.5.1
func (s *slimTrieStringly) Labels(node interface{}) []interface{} {

	n := stNodeID(node)

	rst := []string{}
	labels := s.labels[n]
	for l := range labels {
		rst = append(rst, l)
	}
	sort.Strings(rst)

	r := []interface{}{}
	for _, x := range rst {
		r = append(r, x)
	}
	return r
}

// NodeID implements tree.Tree
//
// Since 0.5.1
func (s *slimTrieStringly) NodeID(node interface{}) string {
	return fmt.Sprintf("%03d", stNodeID(node))
}

// LabelInfo implements tree.Tree
//
// Since 0.5.1
func (s *slimTrieStringly) LabelInfo(label interface{}) string {
	return label.(string)
}

// NodeInfo implements tree.Tree
//
// Since 0.5.1
func (s *slimTrieStringly) NodeInfo(node interface{}) string {
	step, hasStep := s.st.Steps.Get(stNodeID(node))
	if hasStep {
		return fmt.Sprintf("+%d", step)
	}
	return ""
}

// LeafVal implements tree.Tree
//
// Since 0.5.1
func (s *slimTrieStringly) LeafVal(node interface{}) (interface{}, bool) {
	return s.st.Leaves.Get(stNodeID(node))
}

// stNodeID convert a interface to SlimTrie node id.
func stNodeID(node interface{}) int32 {
	if node == nil {
		node = int32(0)
	}
	return node.(int32)
}
