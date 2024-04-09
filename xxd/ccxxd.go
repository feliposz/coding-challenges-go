package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
)

func main() {

	var octestPerLine, grouping, length, seekOffset int
	var displayHelp, littleEndian bool

	flag.IntVar(&octestPerLine, "c", 16, "format <cols> octets per line.")
	flag.IntVar(&grouping, "g", 2, "number of octets per group in normal output.")
	flag.IntVar(&length, "l", math.MaxInt, "stop after <len> octets.")
	flag.IntVar(&seekOffset, "s", 0, "start at <seek> bytes absolute.")
	flag.BoolVar(&displayHelp, "h", false, "print this summary.")
	flag.BoolVar(&littleEndian, "e", false, "little-endian dump.")
	flag.Parse()

	if octestPerLine < 1 || octestPerLine > 256 {
		fmt.Fprintf(os.Stderr, "invalid value for -c (1-256)\n")
		displayHelp = true
	}

	if grouping < 1 || grouping > 32 {
		fmt.Fprintf(os.Stderr, "invalid value for -g (1-32)\n")
		displayHelp = true
	}

	if seekOffset < 0 {
		fmt.Fprintf(os.Stderr, "invalid value for -s (>= 0)\n")
		displayHelp = true
	}

	if displayHelp || !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	var file *os.File
	var err error

	if flag.NArg() > 0 {
		file, err = os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		defer file.Close()
	} else {
		file = os.Stdin
	}

	offset := seekOffset
	_, err = file.Seek(int64(seekOffset), io.SeekStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	buffer := make([]byte, octestPerLine)
	bytesWritten := 0
	for bytesWritten < length {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		bytesLeft := length - bytesWritten
		bytesRead = min(bytesRead, bytesLeft)

		fmt.Printf("%08x:", offset)

		for i := 0; i < octestPerLine; i++ {
			if i%grouping == 0 {
				fmt.Print(" ")
			}
			if i >= bytesRead {
				fmt.Print("  ")
			} else if littleEndian {
				j := (i/grouping)*grouping + grouping - i%grouping - 1
				fmt.Printf("%02x", buffer[j])
			} else {
				fmt.Printf("%02x", buffer[i])
			}
		}

		fmt.Print("  ")

		for i := 0; i < octestPerLine; i++ {
			if i >= bytesRead {
				fmt.Print(" ")
			} else if buffer[i] < 32 || buffer[i] > 127 {
				fmt.Print(".")
			} else {
				fmt.Printf("%c", buffer[i])
			}
		}

		fmt.Print("\n")

		offset += octestPerLine
		bytesWritten += octestPerLine
	}
}
