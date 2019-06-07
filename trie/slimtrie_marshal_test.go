package trie

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/openacid/errors"
	"github.com/openacid/slim/encode"
	"github.com/stretchr/testify/require"
)

func TestSlimTrie_Unmarshal_incompatible(t *testing.T) {

	ta := require.New(t)

	st1, err := NewSlimTrie(encode.Int{}, marshalCase.keys, marshalCase.values)
	ta.Nil(err)

	buf, err := st1.Marshal()
	ta.Nil(err)

	st2, _ := NewSlimTrie(encode.Int{}, nil, nil)

	cases := []struct {
		input string
		want  error
	}{
		// {"1.0.0", nil},
		// {"0.5.8", nil},
		// {"0.5.9", nil},
		// {slimtrieVersion, nil},
		{"0.5.11", ErrIncompatible},
		{"0.6.0", ErrIncompatible},
		{"0.9.9", ErrIncompatible},
		{"1.0.1", ErrIncompatible},
	}

	for i, c := range cases {
		// fmt.Println("load from: ", c.input)
		bad := make([]byte, len(buf))
		copy(bad, buf)
		// clear buf for version
		for i := 0; i < 16; i++ {
			bad[i] = 0
		}
		copy(bad, []byte(c.input))
		err := proto.Unmarshal(bad, st2)
		ta.Equal(c.want, errors.Cause(err), "%d-th: case: %+v", i+1, c)
	}
}

func TestSlimTrie_Unmarshal_old_data(t *testing.T) {

	ta := require.New(t)

	folder := "testdata/"
	finfos, err := ioutil.ReadDir(folder)
	ta.Nil(err)

	for prf, ks := range marshalCases {

		for _, finfo := range finfos {

			fn := finfo.Name()

			if !strings.HasPrefix(fn, prf) {
				continue
			}

			if fn != "slimtrie-data-11-0.5.8" {
				continue
			}

			ver := fn[14:]
			if ver == slimtrieVersion {
				// TODO remove this after 0.5.10 test passed
				continue
			}

			// fmt.Println("load old data:", fn, ver)

			path := filepath.Join(folder, fn)
			b, err := ioutil.ReadFile(path)
			ta.Nil(err)

			st, err := NewSlimTrie(encode.I32{}, nil, nil)
			ta.Nil(err)

			err = proto.Unmarshal(b, st)
			ta.Nil(err)

			// fmt.Println("prf:", prf)
			// fmt.Println("loaded slimtrie:")
			// fmt.Println(st)

			for i, key := range ks {
				v, found := st.Get(key)
				ta.True(found, "search for %v", key)
				ta.Equal(int32(i), v, "search for %v", key)
			}
		}
	}
}
