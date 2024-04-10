package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

var octestPerLine, grouping, length, seekOffset int
var displayHelp, littleEndian, decimalOffset, plainHexDump, reverseOp bool

func init() {
	flag.IntVar(&octestPerLine, "c", 16, "format <cols> octets per line.")
	flag.IntVar(&grouping, "g", 2, "number of octets per group in normal output.")
	flag.IntVar(&length, "l", math.MaxInt, "stop after <len> octets.")
	flag.IntVar(&seekOffset, "s", 0, "start at <seek> bytes absolute.")
	flag.BoolVar(&displayHelp, "h", false, "print this summary.")
	flag.BoolVar(&littleEndian, "e", false, "little-endian dump.")
	flag.BoolVar(&decimalOffset, "d", false, "show offset in decimal instead of hex.")
	flag.BoolVar(&plainHexDump, "p", false, "output in postscript plain hexdump style.")
	flag.BoolVar(&reverseOp, "r", false, "reverse operation: convert (or patch) hexdump into binary.")
}

func main() {
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
		if reverseOp {
			outfile, err = os.OpenFile(flag.Arg(1), os.O_RDWR|os.O_CREATE, 0644)
		} else {
			outfile, err = os.Create(flag.Arg(1))
		}
		if err != nil && !os.IsExist(err) {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		defer outfile.Close()
	}

	if reverseOp {
		reverseDump(infile, outfile)
	} else {
		dump(infile, outfile)
	}
}

func dump(infile, outfile *os.File) {
	offset := 0

	if seekOffset > 0 {
		offset = seekOffset
		_, err := infile.Seek(int64(seekOffset), io.SeekStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
	}

	reader := bufio.NewReaderSize(infile, 1024*1024)
	writer := bufio.NewWriterSize(outfile, 1024*1024)

	if plainHexDump {
		octestPerLine = 30
	}

	buffer := make([]byte, octestPerLine)
	bytesWritten := 0
	for bytesWritten < length {
		bytesRead, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		bytesLeft := length - bytesWritten
		bytesRead = min(bytesRead, bytesLeft)

		if bytesRead == 0 {
			break
		}

		if plainHexDump {

			fmt.Fprintf(writer, "%x\n", buffer[:bytesRead])

		} else {

			if decimalOffset {
				fmt.Fprintf(writer, "%08d:", offset)
			} else {
				fmt.Fprintf(writer, "%08x:", offset)
			}

			for i := 0; i < octestPerLine; i++ {
				if i%grouping == 0 {
					fmt.Fprintf(writer, " ")
				}
				if i >= bytesRead {
					fmt.Fprintf(writer, "  ")
				} else if littleEndian {
					j := (i/grouping)*grouping + grouping - i%grouping - 1
					fmt.Fprintf(writer, "%02x", buffer[j])
				} else {
					fmt.Fprintf(writer, "%02x", buffer[i])
				}
			}

			fmt.Fprintf(writer, "  ")

			for i := 0; i < octestPerLine; i++ {
				if i >= bytesRead {
					fmt.Fprintf(writer, " ")
				} else if buffer[i] < 32 || buffer[i] >= 127 {
					fmt.Fprintf(writer, ".")
				} else {
					fmt.Fprintf(writer, "%c", buffer[i])
				}
			}

			fmt.Fprintf(writer, "\n")
		}

		offset += bytesRead
		bytesWritten += bytesRead
	}

	writer.Flush()
}

func reverseDump(infile, outfile *os.File) {
	if plainHexDump {
		scanner := bufio.NewScanner(infile)
		for scanner.Scan() {
			text := scanner.Text()
			data, err := hex.DecodeString(text)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}
			outfile.Write(data)
		}
	} else {
		reader := bufio.NewReaderSize(infile, 1024*1024)
		base := 16
		if decimalOffset {
			base = 10
		}
		for {
			offsetStr, err := reader.ReadString(':')
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}

			if len(offsetStr) == 0 {
				break
			}

			offsetStr = offsetStr[:len(offsetStr)-1]
			offset, err := strconv.ParseInt(offsetStr, base, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}

			dataStr, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(2)
			}

			data := make([]byte, 0)
			for i := 0; i < len(dataStr)-1; i++ {
				if isSpace(dataStr[i]) && isSpace(dataStr[i+1]) {
					break
				} else if isSpace(dataStr[i]) {
					continue
				} else if isHexDigit(dataStr[i]) && isHexDigit(dataStr[i+1]) {
					data = append(data, decodeHexDigit(dataStr[i])*16+decodeHexDigit(dataStr[i+1]))
					i++
				}
			}

			if len(data) > 0 {
				outfile.Seek(offset, io.SeekStart)
				outfile.Write(data)
			}
		}
	}
}

func isSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}
	return false
}

func isHexDigit(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		return true
	}
	return false
}

func decodeHexDigit(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	panic("invalid hex digit")
}
