package main

import (
	"container/heap"
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

	debugFreqTable := false

	if debugFreqTable {
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

	hnHeap := &HuffNodeHeap{}

	heap.Init(hnHeap)

	for code, count := range freq {
		if count > 0 {
			heap.Push(hnHeap, &HuffNode{count, byte(code), nil, nil})
		}
	}

	for len(*hnHeap) > 1 {
		left := heap.Pop(hnHeap).(*HuffNode)
		right := heap.Pop(hnHeap).(*HuffNode)
		heap.Push(hnHeap, &HuffNode{left.Weight + right.Weight, 0, left, right})
	}

	huffTree := heap.Pop(hnHeap).(*HuffNode)

	debugHuffTree := true
	if debugHuffTree {
		printTree(huffTree, 0)
	}
}

func printTree(node *HuffNode, depth int) {
	if node == nil {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Print("    ")
	}
	if node.Code == 0 {
		fmt.Printf("node weight:%d\n", node.Weight)
	} else {
		printable := node.Code
		if printable < 32 || printable > 127 {
			printable = '?'
		}
		fmt.Printf("char:'%c' code:%02x weight:%d\n", printable, node.Code, node.Weight)
	}
	printTree(node.Left, depth+1)
	printTree(node.Right, depth+1)
}

type HuffNode struct {
	Weight int
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
