package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	if len(data) == 0 {
		invalid()
	}
	tokens, err := tokenize(data)
	if err != nil {
		panic(err)
	}
	for _, token := range tokens {
		fmt.Printf("type=%c", token.Type)
		switch token.Type {
		case 'S':
			fmt.Printf(" content=%q", token.Content)
		case '0':
			fmt.Printf(" value=%f", token.Value)
		}
		fmt.Println()
	}
}

type Token struct {
	Type    byte
	Value   float64
	Content string
}

var ErrKeyWord = errors.New("invalid keyword")
var ErrString = errors.New("invalid string")
var ErrNumber = errors.New("invalid number")
var ErrToken = errors.New("invalid token")

func tokenize(data []byte) (tokens []*Token, err error) {
	for i := 0; i < len(data); i++ {
		if isSpace(data[i]) {
			continue
		}
		token := new(Token)
		switch data[i] {
		case 'n', 't', 'f':
			if i+4 < len(data) && slices.Compare(data[i:i+4], []byte("null")) == 0 {
				token.Type = 'n'
				i += 3
			} else if i+4 < len(data) && slices.Compare(data[i:i+4], []byte("true")) == 0 {
				token.Type = 't'
				i += 3
			} else if i+5 < len(data) && slices.Compare(data[i:i+5], []byte("false")) == 0 {
				token.Type = 'f'
				i += 4
			} else {
				return nil, ErrKeyWord
			}
		case '{', '}', '[', ']', ',', ':':
			token.Type = data[i]
		case '"':
			token.Type = 'S' // string
			j := i + 1
			for j < len(data) && data[j] != '"' {
				if data[j] == '\\' && j < len(data)-1 && data[j+1] == '"' {
					j++ // skip escaped quote
				}
				j++
				// TODO: handle escaping properly
			}
			if data[j] != '"' {
				return nil, ErrString
			}
			token.Content = string(data[i+1 : j])
			i = j
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			token.Type = '0' // number
			j := i + 1
			dot, exponent, expSign := false, false, false
			for j < len(data) {
				if dot && data[i] == '.' {
					return nil, ErrNumber
				} else if data[i] == '.' {
					dot = true
				} else if exponent && data[i] == 'e' || data[i] == 'E' {
					return nil, ErrNumber
				} else if data[i] == 'e' || data[i] == 'E' {
					exponent = true
				} else if !exponent && data[i] == '+' || data[i] == '-' {
					return nil, ErrNumber
				} else if expSign && data[i] == '+' || data[i] == '-' {
					return nil, ErrNumber
				} else if data[i] == '+' || data[i] == '-' {
					expSign = true
				} else if data[j] < '0' || data[j] > '9' {
					return nil, ErrNumber
				}
				j++
			}
			token.Value, err = strconv.ParseFloat(string(data[i+1:j]), 64)
			if err != nil {
				return nil, err
			}
			i = j
		default:
			return nil, ErrToken
		}
		tokens = append(tokens, token)
	}
	return
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func valid() {
	os.Exit(0)
}

func invalid() {
	os.Exit(1)
}
