package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvalidCommand(t *testing.T) {
	input := strings.NewReader("aa\n")

	var output bytes.Buffer
	shell := Shell{
		in:  input,
		out: &output,
	}

	shell.startCli()
	got := output.String()
	expectedResult := "aa: command not found"

	if !strings.Contains(got, expectedResult) {
		t.Errorf("Expected to contain message: %q, got: %q", expectedResult, got)
	}
}

func TestValidCommand(t *testing.T) {
	input := strings.NewReader("echo\n")

	var output bytes.Buffer
	shell := Shell{
		in:  input,
		out: &output,
	}

	shell.startCli()

	got := output.String()

	if strings.Contains(got, "command not found") {
		t.Errorf("Expected command to be found got: %v", got)
	}
}

func TestExitCommand(t *testing.T) {

	input := strings.NewReader("exit 0\n")

	var output bytes.Buffer
	shell := Shell{
		in:  input,
		out: &output,
	}

	exitRequest, exitCode := shell.startCli()

	if exitRequest == false || exitCode != 0 {
		t.Errorf("Expect to exit with code 0 insttead got existRequest: %t, and code: %d", exitRequest, exitCode)
	}
}

func getRawOutput(text string) string {
	withoutDollar := strings.ReplaceAll(text, "$ ", "")

	trimed := strings.TrimSpace(withoutDollar)

	return trimed
}

func TestEchoCommand(t *testing.T) {

	input := strings.NewReader("echo abc def\n")

	var output bytes.Buffer
	shell := Shell{
		in:  input,
		out: &output,
	}

	shell.startCli()
	got := getRawOutput(output.String())
	expectedResult := "abc def"

	if got != expectedResult {
		t.Errorf("Expected result to be %q, got: %q", expectedResult, got)
	}
}

type Case struct {
	input  string
	output string
	name   string
	setup  func() string
	clear  func()
}

func TestTypeCommand(t *testing.T) {
	cases := []Case{
		{input: "type echo", output: "echo is a shell builtin", name: "valid command type echo"},
		{input: "type abc", output: "abc: not found", name: "invalid command type abc"},
		{input: "type script", output: "script is /tmp/123/test2/script", name: "valid not builtin command",
			setup: func() string {
				os.Mkdir("/tmp/123", 0755)
				dir := "/tmp/123"
				fmt.Println(dir)
				os.Setenv("PATH", dir+"/test1"+":"+dir+"/test2")
				os.Mkdir(dir+"/test1", 0755)
				os.Mkdir(dir+"/test2", 0755)
				os.WriteFile(filepath.Join(dir+"/test1", "script"), []byte("fsdfs"), 0644)
				os.WriteFile(filepath.Join(dir+"/test2", "script"), []byte("fsdfs"), 0755)

				return dir
			},
			clear: func() {
				os.RemoveAll("/tmp/123")
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			input := strings.NewReader(testCase.input + "\n")

			var output bytes.Buffer
			shell := Shell{
				in:  input,
				out: &output,
			}

			shell.startCli()
			got := getRawOutput(output.String())

			if got != testCase.output {
				t.Errorf("Expected result to be %q, got: %q", testCase.output, got)
			}

			if testCase.clear != nil {
				testCase.clear()
			}
		})
	}

}
