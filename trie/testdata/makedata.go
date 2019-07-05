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
	trie.MakeMarshaledData("slimtrie-data-11-%s", keys11)
	trie.MakeMarshaledData("slimtrie-data-%s", nil)
}
