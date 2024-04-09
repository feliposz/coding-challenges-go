package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "usage: ccxxd [filename]")
		os.Exit(1)
	}

	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
	defer file.Close()

	bytesPerRow := 16
	grouping := 2
	buffer := make([]byte, bytesPerRow)
	offset := 0
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}

		fmt.Printf("%08x:", offset)

		for i := 0; i < bytesPerRow; i++ {
			if i%grouping == 0 {
				fmt.Print(" ")
			}
			if i >= bytesRead {
				fmt.Print("  ")
			} else {
				fmt.Printf("%02x", buffer[i])
			}
		}

		fmt.Print("  ")

		for i := 0; i < bytesPerRow; i++ {
			if i >= bytesRead {
				fmt.Print(" ")
			} else if buffer[i] < 32 || buffer[i] > 127 {
				fmt.Print(".")
			} else {
				fmt.Printf("%c", buffer[i])
			}
		}

		fmt.Print("\n")

		offset += bytesPerRow
	}
}
