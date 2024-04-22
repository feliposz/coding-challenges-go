package main

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"slices"
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
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	// build character frequency table

	freq := [256]int{}

	for _, b := range data {
		freq[b]++
	}

	debugFreqTable := false

	if debugFreqTable {
		for code, count := range freq {
			fmt.Printf("%6d %c %02x  ", count, toPrintable(byte(code)), code)
			if (code+1)%8 == 0 {
				fmt.Println()
			}
		}
	}

	// build a heap with nodes sorted by frequency count

	hnHeap := &HuffNodeHeap{}

	heap.Init(hnHeap)

	for code, count := range freq {
		if count > 0 {
			heap.Push(hnHeap, &HuffNode{count, true, byte(code), nil, nil})
		}
	}

	// build a binary tree of the nodes
	// https://opendsa-server.cs.vt.edu/ODSA/Books/CS3/html/Huffman.html#building-huffman-coding-trees

	for len(*hnHeap) > 1 {
		left := heap.Pop(hnHeap).(*HuffNode)
		right := heap.Pop(hnHeap).(*HuffNode)
		heap.Push(hnHeap, &HuffNode{left.Weight + right.Weight, false, 0, left, right})
	}

	huffTree := heap.Pop(hnHeap).(*HuffNode)

	debugHuffTree := false
	if debugHuffTree {
		printTree(huffTree, 0)
	}

	prefixCodeTable := [256][]int{}
	buildPrefixCodeTable(huffTree, []int{}, &prefixCodeTable)

	debugPrefixTable := false
	if debugPrefixTable {
		minLen, maxLen := 1000000, 0
		for code, prefix := range prefixCodeTable {
			if len(prefix) == 0 {
				continue
			}
			minLen = min(minLen, len(prefix))
			maxLen = max(maxLen, len(prefix))
			fmt.Printf("'%c' %d %v\n", toPrintable(byte(code)), code, prefix)
		}

		fmt.Printf("minlen: %d\nmaxlen: %d\n", minLen, maxLen)
	}

	originalSize := len(data)
	predictedCompressedSize := 0
	for code, count := range freq {
		predictedCompressedSize += count * len(prefixCodeTable[byte(code)])
	}
	predictedCompressedSize = (predictedCompressedSize + 7) / 8
	fmt.Printf("original size: %d\npredicted compressed size: %d\n", originalSize, predictedCompressedSize)
	fmt.Printf("compression ratio: %f\n", float64(predictedCompressedSize)/float64(originalSize))

	encoded := encodePrefixTable(&prefixCodeTable)
	fmt.Println(len(encoded), encoded)

	testPrefixCodeTable := decodePrefixTable(encoded)

	for i := range prefixCodeTable {
		if slices.Compare(prefixCodeTable[i], testPrefixCodeTable[i]) != 0 {
			fmt.Println(i, prefixCodeTable[i], testPrefixCodeTable[i])
			panic("decoded prefix is different")
		}
	}
}

// encoded format is:
// 1 byte = number of entries (1-256, 0 == 256)
// each entry is:
// 1 byte = character code
// 1 byte = number of bits on the prefix (1-255)
// N bytes = bits for the prefix padded with zeros

func encodePrefixTable(prefixCodeTable *[256][]int) []byte {
	result := []byte{}

	prefixTableSize := 0
	for _, prefix := range prefixCodeTable {
		if len(prefix) > 0 {
			prefixTableSize++
		}
	}
	if prefixTableSize == 0 || prefixTableSize > 256 {
		panic("invalid prefix table size")
	}
	// WARNING: in fact 256 == 0 for our purposes!
	result = append(result, byte(prefixTableSize))
	for code, prefix := range prefixCodeTable {
		bits := len(prefix)
		if bits == 0 {
			continue
		}
		if bits > 255 {
			panic("invalid prefix length")
		}
		result = append(result, byte(code), byte(bits))
		prefixByte := byte(0)
		for i, bit := range prefix {
			prefixByte = prefixByte<<1 | byte(bit)
			switch i + 1 {
			case 8, 16, 24, 32, len(prefix):
				result = append(result, prefixByte)
				prefixByte = 0
			}
		}
	}
	return result
}

func decodePrefixTable(encoded []byte) (result [256][]int) {
	prefixTableSize := int(encoded[0])
	if prefixTableSize == 0 {
		prefixTableSize = 256
	}
	i := 1
	for i < len(encoded) {
		code := encoded[i]
		i++
		bits := encoded[i]
		i++
		fmt.Printf("decoding code: %d bits: %d prefix: ", code, bits)
		prefix := make([]int, 0, bits)
		shift := 8
		for j := 0; j < int(bits); j++ {
			shift--
			prefix = append(prefix, int(encoded[i]>>shift&1))
			if shift == 0 {
				shift = 8
				i++
			}
		}
		fmt.Println(prefix)
		result[code] = prefix
	}
	return
}

func buildPrefixCodeTable(node *HuffNode, prefix []int, prefixCodeTable *[256][]int) {
	if node.IsLeaf {
		prefixCodeTable[node.Code] = prefix
	}
	if node.Left != nil {
		buildPrefixCodeTable(node.Left, append(slices.Clone(prefix), 0), prefixCodeTable)
	}
	if node.Right != nil {
		buildPrefixCodeTable(node.Right, append(slices.Clone(prefix), 1), prefixCodeTable)
	}
}

func printTree(node *HuffNode, depth int) {
	if node == nil {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Print("    ")
	}
	if !node.IsLeaf {
		fmt.Printf("node weight:%d\n", node.Weight)
	} else {
		fmt.Printf("char:'%c' code:%02x weight:%d\n", toPrintable(node.Code), node.Code, node.Weight)
	}
	printTree(node.Left, depth+1)
	printTree(node.Right, depth+1)
}

func toPrintable(ch byte) byte {
	if ch < 32 || ch > 127 {
		return '?'
	}
	return ch
}

type HuffNode struct {
	Weight int
	IsLeaf bool
	Code   byte
	Left   *HuffNode
	Right  *HuffNode
}

type HuffNodeHeap []*HuffNode

func (h HuffNodeHeap) Len() int {
	return len(h)
}

func (h HuffNodeHeap) Less(a, b int) bool {
	return h[a].Weight < h[b].Weight
}

func (h HuffNodeHeap) Swap(a, b int) {
	h[a], h[b] = h[b], h[a]
}

func (h *HuffNodeHeap) Push(x any) {
	*h = append(*h, x.(*HuffNode))
}

func (h *HuffNodeHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
