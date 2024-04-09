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

	if flag.NArg() > 2 {
		fmt.Fprintf(os.Stderr, "expecting input and output filenames only\n")
		displayHelp = true
	}

	if displayHelp || !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	var infile, outfile *os.File
	var err error

	infile, outfile = os.Stdin, os.Stdout

	if flag.NArg() >= 1 {
		infile, err = os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		defer infile.Close()
	}

	if flag.NArg() == 2 {
		outfile, err = os.Create(flag.Arg(1))
		if err != nil && !os.IsExist(err) {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		defer outfile.Close()
	}

	offset := seekOffset
	_, err = infile.Seek(int64(seekOffset), io.SeekStart)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	buffer := make([]byte, octestPerLine)
	bytesWritten := 0
	for bytesWritten < length {
		bytesRead, err := infile.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		bytesLeft := length - bytesWritten
		bytesRead = min(bytesRead, bytesLeft)

		fmt.Fprintf(outfile, "%08x:", offset)

		for i := 0; i < octestPerLine; i++ {
			if i%grouping == 0 {
				fmt.Fprintf(outfile, " ")
			}
			if i >= bytesRead {
				fmt.Fprintf(outfile, "  ")
			} else if littleEndian {
				j := (i/grouping)*grouping + grouping - i%grouping - 1
				fmt.Fprintf(outfile, "%02x", buffer[j])
			} else {
				fmt.Fprintf(outfile, "%02x", buffer[i])
			}
		}

		fmt.Fprintf(outfile, "  ")

		for i := 0; i < octestPerLine; i++ {
			if i >= bytesRead {
				fmt.Fprintf(outfile, " ")
			} else if buffer[i] < 32 || buffer[i] > 127 {
				fmt.Fprintf(outfile, ".")
			} else {
				fmt.Fprintf(outfile, "%c", buffer[i])
			}
		}

		fmt.Fprintf(outfile, "\n")

		offset += octestPerLine
		bytesWritten += octestPerLine
	}
}
