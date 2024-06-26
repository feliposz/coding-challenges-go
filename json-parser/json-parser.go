package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

var payloadOnly bool
var depth, maxDepth int

func main() {

	flag.BoolVar(&payloadOnly, "payload-only", false, "Check if type is object or array")
	flag.IntVar(&maxDepth, "max-depth", math.MaxInt, "Max nesting depth of objects")
	flag.Parse()

	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	var file *os.File
	var err error

	if flag.NArg() == 0 {
		file = os.Stdin
	} else if flag.NArg() == 1 {
		file, err = os.Open(flag.Arg(0))
		if err != nil {
			panic(err)
		}
	} else {
		flag.Usage()
		os.Exit(1)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	result, err := parse(string(data))
	switch err {
	case nil:
		fmt.Printf("%#v\n", result)
	case ErrArray, ErrKeyWord, ErrObject, ErrString, ErrNumber, ErrToken, ErrEmpty, ErrPayload, ErrMaxDepth:
		fmt.Fprintln(os.Stderr, err)
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

var ErrEmpty = errors.New("no data")
var ErrKeyWord = errors.New("invalid keyword")
var ErrString = errors.New("invalid string")
var ErrNumber = errors.New("invalid number")
var ErrToken = errors.New("invalid token")
var ErrArray = errors.New("invalid array")
var ErrObject = errors.New("invalid object")
var ErrPayload = errors.New("invalid payload")
var ErrMaxDepth = errors.New("max depth reached")

func tokenize(data string) (tokens []*Token, err error) {
	var tokenType byte
	var escaping, unicode bool
	unicodeDigits := []rune{}
	content := []rune{}
	for i, c := range data {

		retry := true
		for retry {
			retry = false
			switch tokenType {
			case 0:
				switch c {
				case ' ', '\t', '\n', '\r':
					continue
				case '"':
					tokenType = 'S'
				case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					content = append(content, c)
					tokenType = '0'
				case '[', ']', '{', '}', ',', ':':
					tokens = append(tokens, &Token{byte(c), 0, ""})
				default:
					content = append(content, c)
					tokenType = '*'
				}

			case 'S':
				if escaping {
					if unicode {
						unicodeDigits = append(unicodeDigits, c)
						if len(unicodeDigits) == 4 {
							value, err := strconv.ParseUint(string(unicodeDigits), 16, 32)
							if err != nil {
								fmt.Println("invalid unicode codepoint: ", err)
								return nil, ErrString
							}
							content = append(content, rune(value))
							unicode = false
							escaping = false
							unicodeDigits = unicodeDigits[:0]
						}
						continue
					}
					switch c {
					case '"', '\\', '/':
						content = append(content, c)
						escaping = false
						continue
					case 'r':
						content = append(content, '\r')
						escaping = false
						continue
					case 'n':
						content = append(content, '\n')
						escaping = false
						continue
					case 'b':
						content = append(content, '\b')
						escaping = false
						continue
					case 't':
						content = append(content, '\t')
						escaping = false
						continue
					case 'f':
						content = append(content, '\f')
						escaping = false
						continue
					case 'u':
						unicode = true
						continue
					default:
						return nil, ErrString
					}
				}

				switch c {
				case '"':
					tokens = append(tokens, &Token{tokenType, 0, string(content)})
					content = content[:0]
					tokenType = 0

				case '\\':
					escaping = true

				case '\r', '\n', '\t', '\b':
					return nil, ErrString

				default:
					content = append(content, c)
				}

			case '0':
				parseNumber := false
				switch c {
				case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'e', 'E', '.', '+':
					content = append(content, c)
				default:
					parseNumber = true
				}

				if i == len(data)-1 {
					parseNumber = true
				}

				if parseNumber {
					if len(content) > 1 && content[0] == '0' && content[1] != '.' {
						// No leading zero
						return nil, ErrNumber
					}
					value, err := strconv.ParseFloat(string(content), 64)
					if err != nil {
						return nil, ErrNumber
					}
					tokens = append(tokens, &Token{tokenType, value, ""})
					content = content[:0]
					tokenType = 0
					retry = i < len(data)
				}

			case '*':
				parseKeyword := false
				if c >= 'a' && c <= 'z' {
					content = append(content, c)
				} else {
					parseKeyword = true
				}

				if i == len(data)-1 {
					parseKeyword = true
				}

				if parseKeyword {
					switch string(content) {
					case "null":
						tokenType = 'n'
					case "false":
						tokenType = 'f'
					case "true":
						tokenType = 't'
					default:
						return nil, ErrKeyWord
					}
					tokens = append(tokens, &Token{tokenType, 0, ""})
					content = content[:0]
					tokenType = 0
					retry = i < len(data)
				}
			}
		}
	}
	return tokens, nil
}

func parse(data string) (result interface{}, err error) {
	tokens, err := tokenize(data)
	if err != nil {
		return
	}
	if len(tokens) == 0 {
		return nil, ErrEmpty
	}
	result, end, err := parseTokens(tokens, 0)
	if err != nil {
		return
	}
	if end < len(tokens) {
		return nil, ErrToken
	}
	if payloadOnly {
		switch result.(type) {
		case []interface{}:
			// ok
		case map[string]interface{}:
			// ok
		default:
			return nil, ErrPayload
		}
	}
	return
}

func parseTokens(tokens []*Token, start int) (result interface{}, end int, err error) {
	depth++
	defer func() {
		depth--
	}()
	if depth > maxDepth {
		return nil, 0, ErrMaxDepth
	}
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
		if i < len(tokens) && tokens[i].Type == ',' {
			i++
			if i < len(tokens) && tokens[i].Type == ']' {
				return nil, 0, ErrArray
			}
		}
	}
	if i >= len(tokens) || tokens[i].Type != ']' {
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
	if i >= len(tokens) || tokens[i].Type != '}' {
		return nil, 0, ErrObject
	}
	result = obj
	return
}
