package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode"
)

var countBytes, countLines, countWords, countChars, maxLengths, displayHelp bool

func init() {
	flag.BoolVar(&displayHelp, "h", false, "display this help and exit")
	flag.BoolVar(&displayHelp, "help", false, "display this help and exit")
	flag.BoolVar(&countBytes, "c", false, "print the byte counts")
	flag.BoolVar(&countChars, "m", false, "print the character counts")
	flag.BoolVar(&countLines, "l", false, "print the newline counts")
	flag.BoolVar(&maxLengths, "L", false, "print the maximum display width")
	flag.BoolVar(&countWords, "w", false, "print the word counts")
	flag.BoolVar(&countBytes, "bytes", false, "print the byte counts")
	flag.BoolVar(&countChars, "chars", false, "print the character counts")
	flag.BoolVar(&countLines, "lines", false, "print the newline counts")
	flag.BoolVar(&maxLengths, "max-line-length", false, "print the maximum display width")
	flag.BoolVar(&countWords, "words", false, "print the word counts")
}

type counters struct {
	bytes, lines, words, chars, maxLineLength int64
}

func main() {
	flag.Parse()

	if !flag.Parsed() || displayHelp {
		flag.Usage()
	}

	if !countBytes && !countChars && !countLines && !countWords && !maxLengths {
		countBytes, countLines, countWords = true, true, true
	}

	var fileCount int64
	var total counters

	for _, name := range flag.Args() {
		result := processFilename(name)
		total.lines += result.lines
		total.words += result.words
		total.chars += result.chars
		total.bytes += result.bytes
		total.maxLineLength = max(total.maxLineLength, result.maxLineLength)
		fileCount++
	}

	displayTotals(fileCount, total)

	if fileCount == 0 {
		processFile("", os.Stdin)
	}
}

func processFilename(name string) (result counters) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	result = processFile(name, file)
	file.Close()
	return
}

func processFile(name string, file *os.File) (result counters) {
	if file != os.Stdin {
		stat, err := file.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		result.bytes = stat.Size()
	}

	if file == os.Stdin || countLines || countWords || countChars || maxLengths {
		reader := bufio.NewReader(file)
		if file == os.Stdin {
			result.bytes = 0
		}
		if countWords || countChars || maxLengths {
			for {
				text, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					panic(err)
				}
				prev := ' '
				lineLength := int64(0)
				for _, curr := range text {
					if !unicode.IsSpace(curr) && unicode.IsSpace(prev) {
						result.words++
					}
					prev = curr
					if curr == '\t' {
						// adjust for tab stops
						lineLength = (lineLength + 8) / 8 * 8
					} else if curr != '\n' && curr != '\r' {
						lineLength++
					}
					result.chars++
				}
				result.lines++
				result.bytes += int64(len(text))
				result.maxLineLength = max(result.maxLineLength, lineLength)
			}
		} else {
			// faster, just counting lines
			for {
				_, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					panic(err)
				}
				result.lines++
			}
		}
	}

	if countLines {
		fmt.Printf("%7d ", result.lines)
	}
	if countWords {
		fmt.Printf("%7d ", result.words)
	}
	if countChars {
		fmt.Printf("%7d ", result.chars)
	}
	if countBytes {
		fmt.Printf("%7d ", result.bytes)
	}
	if maxLengths {
		fmt.Printf("%7d ", result.maxLineLength)
	}
	fmt.Printf("%s\n", name)
	return
}

func displayTotals(fileCount int64, total counters) {
	if fileCount > 1 {
		if countLines {
			fmt.Printf("%7d ", total.lines)
		}
		if countWords {
			fmt.Printf("%7d ", total.words)
		}
		if countBytes {
			fmt.Printf("%7d ", total.bytes)
		}
		if countChars {
			fmt.Printf("%7d ", total.chars)
		}
		if maxLengths {
			fmt.Printf("%7d ", total.maxLineLength)
		}
		fmt.Printf("total\n")
	}
}
