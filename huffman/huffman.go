package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: huffman <filename>")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	freq := [256]int{}

	for _, b := range data {
		freq[b]++
	}

	for code, count := range freq {
		printable := code
		if code < 32 || code > 127 {
			printable = '?'
		}
		fmt.Printf("%6d %c %02x  ", count, printable, code)
		if (code+1)%8 == 0 {
			fmt.Println()
		}
	}
}
