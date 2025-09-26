package main

import (
	"reflect"
	"testing"
)

type TestCase struct {
	input          string
	outputCommand  string
	outputArgs     []string
	outputRedirect []string
	pipe           *TestCaseData
}

type TestCaseData struct {
	input          string
	outputCommand  string
	outputArgs     []string
	outputRedirect []string
}

func TestParser(t *testing.T) {

	testCases := []TestCase{
		{
			input:         "echo 'aa bb'",
			outputCommand: "echo",
			outputArgs:    []string{"aa bb"},
		},
		{
			input:         "echo hello   world",
			outputCommand: "echo",
			outputArgs:    []string{"hello", " ", "world"},
		},
		{
			input:         "echo 'example     world' 'hello''script' test''shell",
			outputCommand: "echo",
			outputArgs:    []string{"example     world", " ", "hello", "script", " ", "test", "", "shell"},
		},

		{
			input:          "ls /tmp/baz > /tmp/foo/baz.md",
			outputCommand:  "ls",
			outputArgs:     []string{"/tmp/baz", " "},
			outputRedirect: []string{">", "/tmp/foo/baz.md"},
		},
		{
			input:          "echo 'Hello Maria' 1> /tmp/baz/foo.mdd",
			outputCommand:  "echo",
			outputArgs:     []string{"Hello Maria", " "},
			outputRedirect: []string{"1>", "/tmp/baz/foo.mdd"},
		},
		{
			input:          "echo 'Hello Maria' >> /tmp/baz/foo.mdd",
			outputCommand:  "echo",
			outputArgs:     []string{"Hello Maria", " "},
			outputRedirect: []string{">>", "/tmp/baz/foo.mdd"},
		},
		{
			input:          "echo 'Hello Maria' 1>> /tmp/baz/foo.mdd",
			outputCommand:  "echo",
			outputArgs:     []string{"Hello Maria", " "},
			outputRedirect: []string{"1>>", "/tmp/baz/foo.mdd"},
		},
		{
			input:         "cat /tmp/bar/file-37 | wc",
			outputCommand: "cat",
			outputArgs:    []string{"/tmp/bar/file-37", " "},
			pipe: &TestCaseData{
				outputCommand: "wc",
				outputArgs:    nil,
			},
		},
		{
			input:         "tail -f /tmp/quz/file-19 | head -n 5",
			outputCommand: "tail",
			outputArgs:    []string{"-f", " ", "/tmp/quz/file-19", " "},
			pipe: &TestCaseData{
				outputCommand: "head",
				outputArgs:    []string{"-n", " ", "5"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run("Running input"+testCase.input, func(t *testing.T) {
			parser := NewParser(testCase.input)
			ret, err := parser.parsePipe()

			if err != nil {
				t.Error(err)
			}

			if testCase.outputCommand != string(ret.Command) {
				t.Errorf("Expected command to be: %v, got: %v", testCase.outputCommand, ret.Command)
			}

			if !reflect.DeepEqual(testCase.outputArgs, ret.Arguments) {
				t.Errorf("Expected to got: %#v, insted we have:%#v", testCase.outputArgs, ret.Arguments)
			}
			if testCase.outputRedirect != nil && !reflect.DeepEqual(testCase.outputRedirect, ret.Redirection) {
				t.Errorf("Expected to got: %#v, insted we have:%#v", testCase.outputRedirect, ret.Redirection)

			}

			if testCase.pipe != nil {
				if testCase.pipe.outputCommand != string(ret.pipe[0].Command) {
					t.Errorf("Expected pipe command to be %v, got: %v", testCase.pipe.outputCommand, string(ret.pipe[0].Command))
				}

				if !reflect.DeepEqual(testCase.pipe.outputArgs, ret.pipe[0].Arguments) {
					t.Errorf("Expected to got: %#v, insted we have:%#v", testCase.pipe.outputArgs, ret.pipe[0].Arguments)
				}
			}
		})
	}

}
