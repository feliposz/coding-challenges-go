package main

import (
	"container/heap"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
)

var debugMode bool

func main() {
	var compressMode, decompressMode, showHelp bool

	flag.BoolVar(&compressMode, "c", false, "compress filename to output")
	flag.BoolVar(&decompressMode, "d", false, "decompress filename to output")
	flag.BoolVar(&debugMode, "debug", false, "print lots of debug information")
	flag.BoolVar(&showHelp, "h", false, "display this help information")
	flag.Parse()

	if showHelp || !flag.Parsed() {
		fmt.Fprintln(os.Stderr, "Usage: huffman [-c|-d] <filename> <output>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "Input and output files must be provided")
		os.Exit(1)
	}

	if compressMode && decompressMode {
		fmt.Fprintln(os.Stderr, "Can't apply both compress and decompress!")
		os.Exit(1)
	}

	input, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer input.Close()

	output, err := os.Create(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	defer output.Close()

	if compressMode {
		compress(input, output)
	} else {
		decompress(input, output)
	}
}

func compress(input io.Reader, output io.Writer) {
	data, err := io.ReadAll(input)
	if err != nil {
		panic(err)
	}

	if len(data) == 0 {
		panic("input is empty, nothing to compress")
	}

	// build character frequency table

	freq := make([]int, 256)

	for _, b := range data {
		freq[b]++
	}

	if debugMode {
		fmt.Println("[DEBUG] Frequency table")
		for code, count := range freq {
			fmt.Printf("%6d %c %02x  ", count, toPrintable(byte(code)), code)
			if (code+1)%8 == 0 {
				fmt.Println()
			}
		}
	}

	huffTree := buildHuffmanTreeFromFreq(freq)

	if debugMode {
		fmt.Println("[DEBUG] Huffman Binary Tree")
		printTree(huffTree, 0)
	}

	prefixCodeTable := make([][]int, 256)
	buildPrefixCodeTable(huffTree, []int{}, prefixCodeTable)

	if debugMode {
		fmt.Println("[DEBUG] Prefix Table")
		printPrefixTable(prefixCodeTable)
	}

	originalSize := len(data)
	predictedCompressedSize := 0
	for code, count := range freq {
		predictedCompressedSize += count * len(prefixCodeTable[code])
	}
	predictedCompressedSize = (predictedCompressedSize + 7) / 8
	if debugMode {
		fmt.Println("[DEBUG] Compression size")
		fmt.Printf("original size: %d\npredicted compressed size: %d\n", originalSize, predictedCompressedSize)
		fmt.Printf("compression ratio: %f\n", float64(predictedCompressedSize)/float64(originalSize))
	}

	encodedTable := encodePrefixTable(prefixCodeTable)

	if debugMode {
		fmt.Println("[DEBUG] Decoded table")
		testDecodedTable := decodePrefixTable(encodedTable)

		for i := range prefixCodeTable {
			if slices.Compare(prefixCodeTable[i], testDecodedTable[i]) != 0 {
				fmt.Println(i, prefixCodeTable[i], testDecodedTable[i])
				panic("decoded prefix is different")
			}
		}

		fmt.Println("Encoded and decoded tables match!")
	}

	encodedTableLength := len(encodedTable)
	compressedData := make([]byte, 0, predictedCompressedSize)
	outputByte := byte(0)
	shift := 8
	for _, code := range data {
		for _, bit := range prefixCodeTable[code] {
			shift--
			outputByte |= byte(bit) << shift
			if shift == 0 {
				compressedData = append(compressedData, outputByte)
				outputByte = 0
				shift = 8
			}
		}
	}
	// last byte wasn't complete, so write it out
	if shift != 8 {
		compressedData = append(compressedData, outputByte)
	}

	output.Write([]byte("CCHF")) // header
	output.Write(uint32ToBytes(uint32(len(data))))
	output.Write(uint32ToBytes(uint32(len(compressedData))))
	output.Write(uint32ToBytes(uint32(encodedTableLength)))
	output.Write(encodedTable)
	output.Write(compressedData)
}

func decompress(input io.Reader, output io.Writer) {
	data, err := io.ReadAll(input)
	if err != nil {
		panic(err)
	}

	if string(data[0:4]) != "CCHF" {
		panic("not a valid huffman compressed file")
	}

	decompressedDataLength := bytesToUint32(data[4:8])
	compressedDataLength := bytesToUint32(data[8:12])
	encodedTableLength := bytesToUint32(data[12:16])

	if debugMode {
		fmt.Println("[DEBUG] Header lenghts")
		fmt.Printf("Decompressed Length: %d\nCompressed Length: %d\nEncoded Prefix Table Length: %d\n", decompressedDataLength, compressedDataLength, encodedTableLength)
	}
	encodedTable := data[16 : encodedTableLength+16]

	prefixTable := decodePrefixTable(encodedTable)

	if debugMode {
		fmt.Println("[DEBUG] Decoded prefix table")
		printPrefixTable(prefixTable)
	}

	root := buildHuffmanTreeFromTable(prefixTable)
	if debugMode {
		fmt.Println("[DEBUG] Decoded Huffman binary tree")
		printTree(root, 0)
	}

	outData := make([]byte, 0, decompressedDataLength)

	node := root
outer:
	for i := int(encodedTableLength) + 16; i < len(data); i++ {
		for shift := 7; shift >= 0; shift-- {
			bit := int(data[i] >> shift & 1)
			if bit == 0 {
				node = node.Left
			} else {
				node = node.Right
			}
			if node == nil {
				panic("invalid encoding at offset " + strconv.Itoa(i))
			}
			if node.IsLeaf {
				outData = append(outData, node.Code)
				if len(outData) == int(decompressedDataLength) {
					break outer
				}
				node = root
			}
		}
	}

	output.Write(outData)
}

func uint32ToBytes(x uint32) []byte {
	return []byte{
		byte(x & 0xFF),
		byte((x >> 8) & 0xFF),
		byte((x >> 16) & 0xFF),
		byte((x >> 24) & 0xFF),
	}
}

func bytesToUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func toPrintable(ch byte) byte {
	if ch < 32 || ch > 127 {
		return '?'
	}
	return ch
}

func printPrefixTable(prefixCodeTable [][]int) {
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

// encoded format is:
// 1 byte = number of entries (1-256, 0 == 256)
// each entry is:
// 1 byte = character code
// 1 byte = number of bits on the prefix (1-255)
// N bytes = bits for the prefix padded with zeros

func encodePrefixTable(prefixCodeTable [][]int) []byte {
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
		shift := 8
		for i, bit := range prefix {
			shift--
			prefixByte |= byte(bit) << shift
			if shift == 0 || i == len(prefix)-1 {
				result = append(result, prefixByte)
				prefixByte = 0
				shift = 8
			}
		}
	}
	return result
}

func decodePrefixTable(encoded []byte) (result [][]int) {
	result = make([][]int, 256)

	prefixTableSize := int(encoded[0])
	if prefixTableSize == 0 {
		prefixTableSize = 256
	}
	i := 1
	entries := 0
	for entries < prefixTableSize && i < len(encoded) {
		entries++
		code := encoded[i]
		i++
		bits := encoded[i]
		i++
		// fmt.Printf("decoding code:%d bits:%d prefix:", code, bits)
		prefix := make([]int, 0, bits)
		shift := 8
		for bit := 1; bit <= int(bits); bit++ {
			shift--
			prefix = append(prefix, int(encoded[i]>>shift&1))
			if shift == 0 || bit == int(bits) {
				shift = 8
				i++
			}
		}
		// fmt.Println(prefix)
		result[code] = prefix
	}
	return
}

func buildPrefixCodeTable(node *HuffNode, prefix []int, prefixCodeTable [][]int) {
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

func buildHuffmanTreeFromTable(prefixTable [][]int) *HuffNode {
	root := &HuffNode{0, false, 0, nil, nil}
	for code, prefix := range prefixTable {
		node := root
		for i, bit := range prefix {
			if bit == 0 {
				if node.Left == nil {
					node.Left = &HuffNode{0, false, 0, nil, nil}
				}
				node = node.Left
			} else {
				if node.Right == nil {
					node.Right = &HuffNode{0, false, 0, nil, nil}
				}
				node = node.Right
			}
			if i == len(prefix)-1 {
				node.IsLeaf = true
				node.Code = byte(code)
			}
		}
	}
	return root
}

func buildHuffmanTreeFromFreq(freq []int) *HuffNode {
	// build a heap with nodes sorted by frequency count

	hnHeap := HuffNodeHeap{}

	heap.Init(&hnHeap)

	for code, count := range freq {
		if count > 0 {
			heap.Push(&hnHeap, &HuffNode{count, true, byte(code), nil, nil})
		}
	}

	// special case: data is just a repeated character, make a 1 node + 1 leaf tree
	if len(hnHeap) == 1 {
		left := heap.Pop(&hnHeap).(*HuffNode)
		node := &HuffNode{left.Weight, false, 0, left, nil}
		return node
	}

	// build a binary tree of the nodes
	// https://opendsa-server.cs.vt.edu/ODSA/Books/CS3/html/Huffman.html#building-huffman-coding-trees

	for len(hnHeap) > 1 {
		left := heap.Pop(&hnHeap).(*HuffNode)
		right := heap.Pop(&hnHeap).(*HuffNode)
		heap.Push(&hnHeap, &HuffNode{left.Weight + right.Weight, false, 0, left, right})
	}

	huffTree := heap.Pop(&hnHeap).(*HuffNode)

	return huffTree
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
