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

	prefixCodeTable := map[byte][]int{}
	buildPrefixCodeTable(huffTree, []int{}, prefixCodeTable)

	debugPrefixTable := true
	if debugPrefixTable {
		for code, prefix := range prefixCodeTable {
			fmt.Printf("'%c' %d %v\n", toPrintable(code), code, prefix)
		}

		originalSize := len(data)
		predictedCompressedSize := 0
		for code, count := range freq {
			predictedCompressedSize += count * len(prefixCodeTable[byte(code)])
		}
		predictedCompressedSize = (predictedCompressedSize + 7) / 8
		fmt.Printf("original size: %d\npredicted compressed size: %d\n", originalSize, predictedCompressedSize)
		fmt.Printf("compression ratio: %f\n", float64(predictedCompressedSize)/float64(originalSize))
	}
}

func buildPrefixCodeTable(node *HuffNode, prefix []int, prefixCodeTable map[byte][]int) {
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
