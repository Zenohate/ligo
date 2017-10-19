package ligo

import (
	"fmt"
	"strings"
)

// StripComments function is used to strip the comments from the passed ligo source
func StripComments(ltxt string) string {
	lines := strings.Split(ltxt, "\n")

	inQuotes := false
	final := ""
	for _, line := range lines {
		for _, ch := range line {
			if ch == '"' {
				inQuotes = !inQuotes
				final += string(ch)
				continue
			}
			if ch == ';' && !inQuotes {
				break
			}
			final += string(ch)
		}
		final += "\n"
	}
	return final
}

// ScanTokens is used to get token list from a passed ligo expression.
// This is one of the important functions for parsing a ligo source code.
func ScanTokens(ltxt string) ([]string, error) {
	ltxt = StripComments(ltxt)
	strList := make([]string, 0)
	ltxt = strings.TrimSpace(ltxt)
	if len(ltxt) < 2 {
		return nil, Error("Expected atleast (), got : " + ltxt)
	}

	if ltxt[0] != '(' {
		return nil, Error("Expected '(' at the start of the expression, got : " + string(ltxt[0]) + "\n" +
			ltxt,
		)
	}
	if ltxt[len(ltxt)-1] != ')' {
		return nil, Error("Expected ')' at the end of the expression, got : " + string(ltxt[0]))
	}
	inQuotes := false
	inSBkts := false
	current := ""
	for i := 1; i < len(ltxt); i++ {
		c := string(ltxt[i])
		switch c {
		case " ", "\n", "\r", "\t":
			if !inQuotes && !inSBkts {
				if current == "" {
					continue
				}
				strList = append(strList, current)
				current = ""
				continue
			}
			current += c
		case "|":
			if inSBkts {
				current += c
				inSBkts = false
				continue
			}
			if !inQuotes {
				if current != "" {
					return nil, Error("Closure not separated by a space")
				}
				current += c
				inSBkts = true
				continue
			}
			current += c
		case "\"":
			if !inQuotes {
				inQuotes = true
				if current != "" {
					return nil, Error("Not separated by a space")
				}
				current += "\""
				continue
			}
			current += "\""
			strList = append(strList, current)
			inQuotes = false
			current = ""
		case "[":
			if inSBkts {
				return nil, Error("'[' not expected inside a closure")
			}
			if inQuotes {
				current += c
				continue
			}
			if current != "" {
				return nil, Error("Array not separated by a space")
			}
			off := MatchChars(ltxt, int64(i), '[', ']') + 1
			current = ltxt[i:off]
			strList = append(strList, current)
			i = int(off)
			if strings.TrimSpace(string(ltxt[i])) != "" && strings.TrimSpace(string(ltxt[i])) != ")" {
				return nil, Error("Unexpected character found at array end : " + string(ltxt[i]))
			}
			current = ""
		case "(":
			if inSBkts {
				return nil, Error("'(' not expected inside a closure")
			}
			if inQuotes {
				current += c
				continue
			}
			if current != "" {
				return nil, Error("Expression not separated by a space")
			}
			off := MatchChars(ltxt, int64(i), '(', ')') + 1
			current = ltxt[i:off]
			strList = append(strList, current)
			i = int(off)
			if strings.TrimSpace(string(ltxt[i])) != "" && strings.TrimSpace(string(ltxt[i])) != ")" {
				return nil, Error("Unexpected character found at expression end : " + string(ltxt[i]))
			}
			current = ""
		case ")":
			if inSBkts {
				return nil, Error("')' not expected inside a closure")
			}
			if inQuotes {
				current += c
				continue
			}
			if len(ltxt)-1 != i {
				return nil, Error("Expected EOL, got " + string(ltxt[i]) + " at " + fmt.Sprint(i))
			}
			if current != "" {
				strList = append(strList, current)
				current = ""
			}
		case "]":
			if inSBkts {
				return nil, Error("']' not expected inside a closure")
			}
			if inQuotes {
				current += c
				continue
			}
			strList = append(strList, current)
			current = ""
		default:
			current += c
		}
	}
	if inQuotes {
		return nil, Error("Quote not closed correctly")
	}
	if inSBkts {
		return nil, Error("Closure not closed correctly")
	}
	return strList, nil
}

// MatchChars function is used to return the offset at which the matching character of the passed character
// is found in the passed string. Generally used to match brackets.
func MatchChars(ltxt string, off int64, open byte, close byte) int64 {
	if int64(len(ltxt)) <= off {
		return -1
	}
	if ltxt[off] != open {
		return -1
	}
	count := 1
	inQuotes := false
	for i := off + 1; i < int64(len(ltxt)); i++ {
		if ltxt[i] == '"' {
			inQuotes = !inQuotes
		}
		if ltxt[i] == open && !inQuotes {
			count++
		}
		if ltxt[i] == close && !inQuotes {
			count--
		}
		if count == 0 {
			return i
		}
	}
	return -1
}

// getVarsFromClosure is used to extract all the parameter names from a
// closure of a function definition in ligo
// (ie., "|a b v r|" yields an array containing "a", "b", "v" and "r")
func getVarsFromClosure(cl string) []string {
	current := ""
	retParams := make([]string, 0)
	for _, val := range cl {
		if strings.TrimSpace(string(val)) == "" || val == '|' {
			if current == "" {
				continue
			}
			retParams = append(retParams, current)
			current = ""
			continue
		}
		current += string(val)
	}
	return retParams
}

// isVariate is used to check whether a given token string is passed as a variate parameter.
func isVariate(str string) bool {
	if len(str) > 4 && str[:3] == "..." && str[3] != '.' {
		return true
	}
	return false
}
