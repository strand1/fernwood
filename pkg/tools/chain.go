package tools

import (
	"strings"
)

// Operator represents a command chaining operator
type Operator int

const (
	OpNone Operator = iota
	OpAnd   // &&
	OpOr    // ||
	OpSeq   // ;
	OpPipe  // |
)

// Segment represents a single command in a chain
type Segment struct {
	Command   string
	Operator  Operator
	StdinData string
}

// ParseChain splits a command string into segments based on operators
// while respecting quoted strings and escape sequences
func ParseChain(input string) []Segment {
	if input == "" {
		return nil
	}

	segments := make([]Segment, 0)
	tokens := tokenize(input)
	
	if len(tokens) == 0 {
		return nil
	}

	currentCommand := ""
	operator := OpNone
	
	for i, token := range tokens {
		// Check if this token is an operator
		if op := parseOperator(token); op != OpNone {
			// We have a complete command, add it to segments
			if currentCommand != "" {
				segments = append(segments, Segment{
					Command:  strings.TrimSpace(currentCommand),
					Operator: operator,
				})
				currentCommand = ""
			}
			operator = op
		} else {
			// This is part of a command
			if currentCommand != "" {
				currentCommand += " "
			}
			currentCommand += token
		}
		
		// Handle the last command
		if i == len(tokens)-1 && currentCommand != "" {
			segments = append(segments, Segment{
				Command:  strings.TrimSpace(currentCommand),
				Operator: operator,
			})
		}
	}
	
	return segments
}

// tokenize splits input into tokens while respecting quotes and escape sequences
func tokenize(input string) []string {
	var tokens []string
	var currentToken strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false
	
	for _, r := range input {
		if escaped {
			currentToken.WriteRune(r)
			escaped = false
			continue
		}
		
		switch r {
		case '\\':
			if inSingleQuote {
				currentToken.WriteRune(r)
			} else {
				escaped = true
			}
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
				currentToken.WriteRune(r)
			} else {
				currentToken.WriteRune(r)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
				currentToken.WriteRune(r)
			} else {
				currentToken.WriteRune(r)
			}
		case ' ', '\t':
			if !inSingleQuote && !inDoubleQuote {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
			} else {
				currentToken.WriteRune(r)
			}
		case '&':
			if !inSingleQuote && !inDoubleQuote {
				// Check for &&
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
				tokens = append(tokens, "&")
			} else {
				currentToken.WriteRune(r)
			}
		case '|':
			if !inSingleQuote && !inDoubleQuote {
				// Check for ||
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
				tokens = append(tokens, "|")
			} else {
				currentToken.WriteRune(r)
			}
		case ';':
			if !inSingleQuote && !inDoubleQuote {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
				tokens = append(tokens, ";")
			} else {
				currentToken.WriteRune(r)
			}
		default:
			currentToken.WriteRune(r)
		}
	}
	
	// Add the last token
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	
	// Combine operators
	combinedTokens := make([]string, 0)
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "&" && i+1 < len(tokens) && tokens[i+1] == "&" {
			combinedTokens = append(combinedTokens, "&&")
			i++ // Skip next token
		} else if tokens[i] == "|" && i+1 < len(tokens) && tokens[i+1] == "|" {
			combinedTokens = append(combinedTokens, "||")
			i++ // Skip next token
		} else {
			combinedTokens = append(combinedTokens, tokens[i])
		}
	}
	
	return combinedTokens
}

// parseOperator converts a string to an Operator
func parseOperator(token string) Operator {
	switch token {
	case "&&":
		return OpAnd
	case "||":
		return OpOr
	case ";":
		return OpSeq
	case "|":
		return OpPipe
	default:
		return OpNone
	}
}

// String returns a string representation of an Operator
func (op Operator) String() string {
	switch op {
	case OpAnd:
		return "&&"
	case OpOr:
		return "||"
	case OpSeq:
		return ";"
	case OpPipe:
		return "|"
	default:
		return ""
	}
}