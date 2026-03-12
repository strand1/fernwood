package tools

import (
	"reflect"
	"testing"
)

func TestParseChain(t *testing.T) {
	tests := []struct {
		input    string
		expected []Segment
	}{
		{
			input:    "",
			expected: nil,
		},
		{
			input: "ls",
			expected: []Segment{
				{Command: "ls", Operator: OpNone},
			},
		},
		{
			input: "ls && echo hello",
			expected: []Segment{
				{Command: "ls", Operator: OpNone},
				{Command: "echo hello", Operator: OpAnd},
			},
		},
		{
			input: "cat file.txt | grep error",
			expected: []Segment{
				{Command: "cat file.txt", Operator: OpNone},
				{Command: "grep error", Operator: OpPipe},
			},
		},
		{
			input: "ls ; pwd ; whoami",
			expected: []Segment{
				{Command: "ls", Operator: OpNone},
				{Command: "pwd", Operator: OpSeq},
				{Command: "whoami", Operator: OpSeq},
			},
		},
		{
			input: "cmd1 || cmd2 && cmd3",
			expected: []Segment{
				{Command: "cmd1", Operator: OpNone},
				{Command: "cmd2", Operator: OpOr},
				{Command: "cmd3", Operator: OpAnd},
			},
		},
	}

	for _, test := range tests {
		result := ParseChain(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("ParseChain(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "ls",
			expected: []string{"ls"},
		},
		{
			input:    "ls -l",
			expected: []string{"ls", "-l"},
		},
		{
			input:    "echo 'hello world'",
			expected: []string{"echo", "'hello world'"},
		},
		{
			input:    `echo "hello world"`,
			expected: []string{"echo", `"hello world"`},
		},
		{
			input:    "cmd1 && cmd2",
			expected: []string{"cmd1", "&&", "cmd2"},
		},
		{
			input:    "cmd1 || cmd2",
			expected: []string{"cmd1", "||", "cmd2"},
		},
		{
			input:    "cmd1 ; cmd2",
			expected: []string{"cmd1", ";", "cmd2"},
		},
		{
			input:    "cmd1 | cmd2",
			expected: []string{"cmd1", "|", "cmd2"},
		},
	}

	for _, test := range tests {
		result := tokenize(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("tokenize(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestParseOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected Operator
	}{
		{"", OpNone},
		{"&&", OpAnd},
		{"||", OpOr},
		{";", OpSeq},
		{"|", OpPipe},
		{"ls", OpNone},
		{"echo", OpNone},
	}

	for _, test := range tests {
		result := parseOperator(test.input)
		if result != test.expected {
			t.Errorf("parseOperator(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestOperatorString(t *testing.T) {
	tests := []struct {
		input    Operator
		expected string
	}{
		{OpNone, ""},
		{OpAnd, "&&"},
		{OpOr, "||"},
		{OpSeq, ";"},
		{OpPipe, "|"},
	}

	for _, test := range tests {
		result := test.input.String()
		if result != test.expected {
			t.Errorf("Operator(%v).String() = %q; expected %q", test.input, result, test.expected)
		}
	}
}