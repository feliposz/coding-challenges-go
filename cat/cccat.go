package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {

	names := os.Args[1:]

	if len(names) == 0 {
		names = append(names, "-")
	}

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
			fmt.Print(text)
		}
	}
}
