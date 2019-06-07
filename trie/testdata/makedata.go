package main

import "github.com/openacid/slim/trie"

var (
	keys11 = []string{
		"abc",
		"abcd",
		"abcdx",
		"abcdy",
		"abcdz",
		"abd",
		"abde",
		"bc",
		"bcd",
		"bcde",
		"cde",
	}
)

func main() {
	for prf, ks := range trie.marshalCases {
		trie.MakeMarshaledData(pref+"%s", ks)
	}
}
