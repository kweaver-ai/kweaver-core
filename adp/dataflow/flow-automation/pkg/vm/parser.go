package vm

import (
	"fmt"
	"strconv"
	"unicode"
)

type TokenType uint

const (
	TokenLiteral  TokenType = 1
	TokenVariable TokenType = 2
)

type Token struct {
	Type       TokenType
	Value      interface{}
	AccessList []Token
}

func Parse(s string) []Token {
	chars := []rune(s)
	var tokens []Token
	currentPos := 0
	textBuffer := ""

	for currentPos < len(chars) {
		c := chars[currentPos]
		if c != '{' {
			textBuffer += string(c)
			currentPos++
			continue
		}

		if len(textBuffer) > 0 {
			tokens = appendToken(tokens, Token{Type: TokenLiteral, Value: textBuffer})
			textBuffer = ""
		}

		if currentPos+1 < len(chars) && chars[currentPos+1] == '{' {
			if currentPos+2 < len(chars) && chars[currentPos+2] == '{' {
				start := currentPos
				currentPos += 3

				found := false
				end := currentPos
				for currentPos < len(chars) {
					if currentPos+2 < len(chars) &&
						chars[currentPos] == '}' &&
						chars[currentPos+1] == '}' &&
						chars[currentPos+2] == '}' {
						found = true
						end = currentPos
						currentPos += 3
						break
					}
					currentPos++
				}

				if found {
					content := chars[start+3 : end]
					tokens = appendToken(tokens, Token{
						Type:  TokenLiteral,
						Value: string(content),
					})
				} else {
					textBuffer += string(chars[start])
					currentPos = start + 1
					continue
				}
			} else {
				start := currentPos
				currentPos += 2

				found := false
				end := currentPos
				for currentPos < len(chars) {
					if currentPos+1 < len(chars) &&
						chars[currentPos] == '}' &&
						chars[currentPos+1] == '}' {
						found = true
						end = currentPos
						currentPos += 2
						break
					}
					currentPos++
				}

				if found {
					content := chars[start+2 : end]
					varToken := parseVariable(trimSpaceRune(content))
					tokens = appendToken(tokens, varToken)
				} else {
					text := string(chars[start:currentPos])
					tokens = appendToken(tokens, Token{Type: TokenLiteral, Value: text})
				}
			}
		} else {
			textBuffer += string(c)
			currentPos++
		}
	}

	if len(textBuffer) > 0 {
		tokens = appendToken(tokens, Token{Type: TokenLiteral, Value: textBuffer})
	}

	return tokens
}

func trimSpaceRune(runes []rune) []rune {
	start := 0
	for start < len(runes) && unicode.IsSpace(runes[start]) {
		start++
	}

	end := len(runes) - 1
	for end >= 0 && unicode.IsSpace(runes[end]) {
		end--
	}

	if start > end {
		return []rune{}
	}

	return runes[start : end+1]
}

func appendToken(tokens []Token, token Token) []Token {
	if l := len(tokens); l > 0 && token.Type == TokenLiteral && tokens[l-1].Type == TokenLiteral {
		tokens[l-1].Value = fmt.Sprintf("%v%v", tokens[l-1].Value, token.Value)
		return tokens
	}

	return append(tokens, token)
}

func parseVariable(chars []rune) Token {
	parts := splitAccessChain(chars)
	if len(parts) == 0 {
		return Token{Type: TokenVariable, Value: ""}
	}
	varToken := Token{
		Type:  TokenVariable,
		Value: parts[0],
	}
	for _, part := range parts[1:] {
		lit := parseAccessPart(part)
		varToken.AccessList = append(varToken.AccessList, lit)
	}
	return varToken
}

func splitAccessChain(chars []rune) []string {
	parts := []string{}
	current := ""
	inQuotes := false
	var quoteChar rune = 0

	for _, c := range chars {
		if inQuotes {
			if c == rune(quoteChar) {
				inQuotes = false
				quoteChar = 0
				current += string(c)
			} else {
				current += string(c)
			}
			continue
		}
		if c == '.' {
			parts = append(parts, current)
			current = ""
			continue
		}
		if c == '"' || c == '\'' {
			inQuotes = true
			quoteChar = c
			current += string(c)
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func parseAccessPart(part string) Token {
	stripped := part
	if len(part) >= 2 && (part[0] == '"' || part[0] == '\'') && part[0] == part[len(part)-1] {
		stripped = part[1 : len(part)-1]
		return Token{Type: TokenLiteral, Value: stripped}
	}

	if num, err := strconv.Atoi(stripped); err == nil {
		return Token{Type: TokenLiteral, Value: num}
	}
	return Token{Type: TokenLiteral, Value: stripped}
}
