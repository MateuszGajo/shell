package main

import (
	"reflect"
	"testing"
)

type TestCase struct {
	input         string
	outputCommand string
	outputArgs    []string
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
			outputArgs:    []string{"hello", "world"},
		},
		{
			input:         "echo 'example     world' 'hello''script' test''shell",
			outputCommand: "echo",
			outputArgs:    []string{"example     world", "helloscript", "testshell"},
		},
	}

	for _, testCase := range testCases {
		t.Run("Running input"+testCase.input, func(t *testing.T) {
			parser := NewParser(testCase.input)
			ret, err := parser.ParseCommand()

			if err != nil {
				t.Error(err)
			}

			if testCase.outputCommand != string(ret.Command) {
				t.Errorf("Expected command to be: %v, got: %v", testCase.outputCommand, ret.Command)
			}

			if !reflect.DeepEqual(testCase.outputArgs, ret.Arguments) {
				t.Errorf("Expected to got: %#v, insted we have:%#v", testCase.outputArgs, ret.Arguments)
			}
		})
	}

	// args = "echo 'example     world' 'hello''script' test''shell"

	// parser = NewParser(args)
	// ret = parser.parseInput()
	// expected := []string{"echo", "example     world", "helloscript", "testshell"}

	// if !reflect.DeepEqual(ret, expected) {
	// 	t.Errorf("Expected: %#v, got:  %#v", expected, ret)
	// }

	// args = "echo shell     world"

	// parser = NewParser(args)
	// ret = parser.parseInput()

	// expected = []string{"echo", "shell", "world"}

	// if !reflect.DeepEqual(ret, expected) {
	// 	t.Errorf("Expected: %#v, got:  %#v", expected, ret)
	// }
}
