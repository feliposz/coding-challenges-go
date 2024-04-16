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

	result, err := parse(data)
	switch err {
	case nil:
		fmt.Printf("%#v\n", result)
	case ErrArray, ErrKeyWord, ErrObject, ErrString, ErrNumber, ErrToken, ErrEmpty:
		fmt.Println(err)
		os.Exit(1)
	default:
		panic(err)
	}
}

type Token struct {
	Type    byte
	Value   float64
	Content string
}

var ErrEmpty = errors.New("invalid keyword")
var ErrKeyWord = errors.New("invalid keyword")
var ErrString = errors.New("invalid string")
var ErrNumber = errors.New("invalid number")
var ErrToken = errors.New("invalid token")
var ErrArray = errors.New("invalid array")
var ErrObject = errors.New("invalid object")

func tokenize(data []byte) (tokens []*Token, err error) {
	for i := 0; i < len(data); i++ {
		if isSpace(data[i]) {
			continue
		}
		token := new(Token)
		switch data[i] {
		case 'n', 't', 'f':
			if i+4 <= len(data) && slices.Compare(data[i:i+4], []byte("null")) == 0 {
				token.Type = 'n'
				i += 3
			} else if i+4 <= len(data) && slices.Compare(data[i:i+4], []byte("true")) == 0 {
				token.Type = 't'
				i += 3
			} else if i+5 <= len(data) && slices.Compare(data[i:i+5], []byte("false")) == 0 {
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
				if dot && data[j] == '.' {
					return nil, ErrNumber
				} else if data[j] == '.' {
					dot = true
				} else if exponent && (data[j] == 'e' || data[j] == 'E') {
					return nil, ErrNumber
				} else if data[j] == 'e' || data[j] == 'E' {
					exponent = true
				} else if !exponent && (data[j] == '+' || data[j] == '-') {
					return nil, ErrNumber
				} else if expSign && (data[j] == '+' || data[j] == '-') {
					return nil, ErrNumber
				} else if exponent && (data[j] == '+' || data[j] == '-') {
					expSign = true
				} else if data[j] < '0' || data[j] > '9' {
					break
				}
				j++
			}
			token.Value, err = strconv.ParseFloat(string(data[i:j]), 64)
			if err != nil {
				return nil, err
			}
			i = j - 1
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

func parse(data []byte) (result interface{}, err error) {
	tokens, err := tokenize(data)
	if err != nil {
		return
	}
	if len(tokens) == 0 {
		return nil, ErrEmpty
	}
	//printTokens(tokens)
	result, _, err = parseTokens(tokens, 0)
	return
}

func printTokens(tokens []*Token) {
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

func parseTokens(tokens []*Token, start int) (result interface{}, end int, err error) {
	switch tokens[start].Type {
	case '[':
		return parseArray(tokens, start)
	case '{':
		return parseObject(tokens, start)
	case 'S':
		return tokens[start].Content, start + 1, nil
	case '0':
		return tokens[start].Value, start + 1, nil
	case 'n':
		return nil, start + 1, nil
	case 't':
		return true, start + 1, nil
	case 'f':
		return false, start + 1, nil
	default:
		return nil, 0, ErrToken
	}
}

func parseArray(tokens []*Token, start int) (result interface{}, end int, err error) {
	arr := make([]interface{}, 0)
	i := start + 1
	for i < len(tokens) {
		if tokens[i].Type == ']' {
			end = i + 1
			break
		}
		var value interface{}
		value, i, err = parseTokens(tokens, i)
		if err != nil {
			return nil, 0, err
		}
		arr = append(arr, value)
		if tokens[i].Type == ',' {
			i++
			if tokens[i].Type == ']' {
				return nil, 0, ErrArray
			}
			continue
		}
	}
	if tokens[i].Type != ']' {
		return nil, 0, ErrArray
	}
	result = arr
	return
}

func parseObject(tokens []*Token, start int) (result interface{}, end int, err error) {
	obj := make(map[string]interface{})
	i := start + 1
	for i < len(tokens) {
		if tokens[i].Type == '}' {
			end = i + 1
			break
		}
		var key string
		var value interface{}
		if tokens[i].Type != 'S' {
			return nil, 0, ErrObject
		}
		key = tokens[i].Content
		i++
		if tokens[i].Type != ':' {
			return nil, 0, ErrObject
		}
		i++
		value, i, err = parseTokens(tokens, i)
		if err != nil {
			return nil, 0, err
		}
		obj[key] = value
		if i < len(tokens) && tokens[i].Type == ',' {
			i++
			if i < len(tokens) && tokens[i].Type == '}' {
				return nil, 0, ErrObject
			}
		}
	}
	if tokens[i].Type != '}' {
		return nil, 0, ErrObject
	}
	result = obj
	return
}
