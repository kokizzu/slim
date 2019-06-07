package trie

import "github.com/openacid/low/bitmap"

func (b *Bitmap) Extend(n int32) {
	b.Words = bitmap.Extend(b.Words, n)

	if b.RankIndex != nil {
		b.RankIndex = bitmap.IndexRank64(b.Words)
	}
}

func (b *Bitmap) Has(i int32) bool {
	return bitmap.Get(b.Words, i) != 0
}

func (b *Bitmap) ToArray() []int32 {
	return bitmap.ToArray(b.Words)
}

func (b *Bitmap) Cnt() int32 {
	_, r := bitmap.Rank64(b.Words, b.RankIndex, int32(len(b.Words)*64-1))
	return r
}
