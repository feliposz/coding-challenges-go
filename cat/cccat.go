package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var printLineNumbers, printNonBlankLineNumbers, showEnds, showTabs, showNonPrint, showAll, showNonPrintAndEnds, showNonPrintAndTabs, displayHelp bool

func init() {
	flag.BoolVar(&printLineNumbers, "n", false, "number all output lines")
	flag.BoolVar(&printLineNumbers, "number", false, "number all output lines")
	flag.BoolVar(&printNonBlankLineNumbers, "b", false, "number nonempty output lines, overrides -n")
	flag.BoolVar(&printNonBlankLineNumbers, "number-nonblank", false, "number nonempty output lines, overrides -n")
	flag.BoolVar(&showEnds, "E", false, "display $ at end of each line")
	flag.BoolVar(&showEnds, "show-ends", false, "display $ at end of each line")
	flag.BoolVar(&showNonPrint, "v", false, "use ^ and M- notation, except for LFD and TAB")
	flag.BoolVar(&showNonPrint, "show-nonprinting", false, "use ^ and M- notation, except for LFD and TAB")
	flag.BoolVar(&showTabs, "T", false, "display TAB characters as ^I")
	flag.BoolVar(&showTabs, "show-tabs", false, "display TAB characters as ^I")
	flag.BoolVar(&showAll, "A", false, "equivalent to -vET")
	flag.BoolVar(&showAll, "show-all", false, "equivalent to -vET")
	flag.BoolVar(&showNonPrintAndEnds, "e", false, "equivalent to -vE")
	flag.BoolVar(&showNonPrintAndTabs, "t", false, "equivalent to -vT")
	flag.BoolVar(&displayHelp, "h", false, "display this help and exit")
	flag.BoolVar(&displayHelp, "help", false, "display this help and exit")
}

func main() {
	flag.Parse()

	if displayHelp || !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	if showAll {
		showEnds, showTabs, showNonPrint = true, true, true
	}
	if showNonPrintAndEnds {
		showEnds, showNonPrint = true, true
	}
	if showNonPrintAndTabs {
		showTabs, showNonPrint = true, true
	}

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

		reader := bufio.NewReaderSize(file, 1024*1024)

		for {
			text, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					panic(err)
				} else if len(text) == 0 {
					break
				}
			}
			if printNonBlankLineNumbers {
				if len(strings.TrimRight(string(text), "\r\n")) > 0 {
					lineNumber++
					fmt.Printf("%6d\t", lineNumber)
				}
			} else if printLineNumbers {
				lineNumber++
				fmt.Printf("%6d\t", lineNumber)
			}
			if showEnds || showTabs || showNonPrint {
				for _, c := range text {
					if showEnds && c == '\n' {
						fmt.Print("$\n")
					} else if showTabs && c == '\t' {
						fmt.Print("^I")
					} else if showNonPrint && (c < 32 || c >= 127) {
						switch {
						case c == 0:
							fmt.Print("^@")
						case c == '\n':
							fmt.Print("\n")
						case c < 32:
							fmt.Printf("^%c", c-1+'A')
						case c == 127:
							fmt.Print("^?")
						case c == 128:
							fmt.Print("M-^@")
						case c < 160:
							fmt.Printf("M-^%c", c-129+'A')
						case c < 255:
							fmt.Printf("M-%c", c-128)
						case c == 255:
							fmt.Print("M-^?")
						}
					} else {
						fmt.Printf("%c", c)
					}
				}
			} else {
				fmt.Print(string(text))
			}
		}
	}
}
