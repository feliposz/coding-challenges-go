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

func main() {
	flag.Parse()

	if !flag.Parsed() || displayHelp {
		flag.Usage()
	}

	if !countBytes && !countChars && !countLines && !countWords && !maxLengths {
		countBytes, countLines, countWords = true, true, true
	}

	var fileCount, totalBytes, totalLines, totalWords, totalChars, totalMaxLineLengths int64

	for _, name := range flag.Args() {
		bytes, lines, words, chars, maxLineLength := processFile(name)
		totalLines += lines
		totalWords += words
		totalChars += chars
		totalBytes += bytes
		totalMaxLineLengths = max(totalMaxLineLengths, maxLineLength)
		fileCount++
	}

	displayTotals(fileCount, totalLines, totalWords, totalBytes, totalChars, totalMaxLineLengths)
}

func processFile(name string) (bytes, lines, words, chars, maxLineLength int64) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	bytes = stat.Size()

	if countLines || countWords || countChars || maxLengths {
		reader := bufio.NewReader(file)
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
						words++
					}
					prev = curr
					if curr == '\t' {
						// adjust for tab stops
						lineLength = (lineLength + 8) / 8 * 8
					} else if curr != '\n' && curr != '\r' {
						lineLength++
					}
					chars++
				}
				lines++
				maxLineLength = max(maxLineLength, lineLength)
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
				lines++
			}
		}
	}

	if countLines {
		fmt.Printf("%6d ", lines)
	}
	if countWords {
		fmt.Printf("%6d ", words)
	}
	if countChars {
		fmt.Printf("%6d ", chars)
	}
	if countBytes {
		fmt.Printf("%6d ", bytes)
	}
	if maxLengths {
		fmt.Printf("%6d ", maxLineLength)
	}
	fmt.Printf("%s\n", name)
	return
}

func displayTotals(fileCount, totalLines, totalWords, totalBytes, totalChars, totalMaxLineLengths int64) {
	if fileCount > 1 {
		if countLines {
			fmt.Printf("%6d ", totalLines)
		}
		if countWords {
			fmt.Printf("%6d ", totalWords)
		}
		if countBytes {
			fmt.Printf("%6d ", totalBytes)
		}
		if countChars {
			fmt.Printf("%6d ", totalChars)
		}
		if maxLengths {
			fmt.Printf("%6d ", totalMaxLineLengths)
		}
		fmt.Printf("total\n")
	}
}
