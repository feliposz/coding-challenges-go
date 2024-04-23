package main

import (
	"bytes"
	"slices"
	"testing"
)

func TestCompressAndDecompress(t *testing.T) {

	// simple strings
	testCases := [][]byte{
		[]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		[]byte("hello huffman"),
		[]byte("aaabbcaaabbcaaabbcaaabbc"),
		[]byte("Coding\tChallenges\nAre\tFun!!! :)"),
	}

	// "complex" binary data
	tmp := []byte{}
	for i := 0; i < 256; i++ {
		for j := 0; j <= i; j++ {
			tmp = append(tmp, byte(j))
		}
	}

	testCases = append(testCases, tmp)

	for i, original := range testCases {

		compressed := new(bytes.Buffer)
		decompressed := new(bytes.Buffer)

		compress(bytes.NewReader(original), compressed)
		decompress(compressed, decompressed)

		if slices.Compare(original, decompressed.Bytes()) != 0 {
			t.Errorf("decompression test #%d failed. want: %q - got: %q\n", i, original, decompressed.Bytes())
		}
	}
}
