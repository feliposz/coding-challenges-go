package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var printLineNumbers, printNonBlankLineNumbers bool

func init() {
	flag.BoolVar(&printLineNumbers, "n", false, "number all output lines")
	flag.BoolVar(&printLineNumbers, "number", false, "number all output lines")
	flag.BoolVar(&printNonBlankLineNumbers, "b", false, "number nonempty output lines, overrides -n")
	flag.BoolVar(&printNonBlankLineNumbers, "number-nonblank", false, "number nonempty output lines, overrides -n")
}

func main() {
	flag.Parse()

	names := flag.Args()

	if len(names) == 0 {
		names = append(names, "-")
	}

	lineNumber := 0

	for _, name := range names {
		var file *os.File
		var err error
		if name == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			defer file.Close()
		}

		reader := bufio.NewReader(file)

		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			if printNonBlankLineNumbers {
				if len(strings.TrimRight(text, "\r\n")) > 0 {
					lineNumber++
					fmt.Printf("%6d\t", lineNumber)
				}
			} else if printLineNumbers {
				lineNumber++
				fmt.Printf("%6d\t", lineNumber)
			}
			fmt.Print(text)
		}
	}
}
