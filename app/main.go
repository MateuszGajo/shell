package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Command string

const (
	EchoCommand Command = "echo"
	ExitCommand Command = "exit"
	TypeCommand Command = "type"
	PwdCommand  Command = "pwd"
	CdCommand   Command = "cd"
)

var builtinCommands = []Command{EchoCommand, ExitCommand, TypeCommand, PwdCommand, CdCommand}

type Shell struct {
	in        io.Reader
	stdout    io.Writer
	stderr    io.Writer
	directory string
}

func isBuiltinCommand(command Command) bool {
	for _, item := range builtinCommands {
		if item == command {
			return true
		}
	}
	return false
}

func findFile(directories string, file string) (bool, string) {
	directoriesSplit := strings.Split(directories, ":")

	for _, item := range directoriesSplit {
		path := item + "/" + file
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() && info.Mode().Perm()&0100 != 0 {
			return true, path
		}

	}
	return false, ""

}

func (shell *Shell) handleEchoCommand(args []string) (string, error) {
	return strings.Join(args, ""), nil
}

func (shell *Shell) handleTypeCommand(args []string) (string, error) {

	if len(args) > 1 {
		return "", fmt.Errorf("expected only one argument")
	}

	if isBuiltinCommand(Command(args[0])) {
		return args[0] + " is a shell builtin", nil
	} else if ok, path := findFile(os.Getenv("PATH"), args[0]); ok {
		return args[0] + " is " + path, nil
	} else {
		return "", fmt.Errorf("%s", args[0]+": not found")
	}

}

func (shell *Shell) handleExternalCommand(command Command, args []string) (string, error) {
	ok, _ := findFile(os.Getenv("PATH"), string(command))

	if !ok {
		return "", fmt.Errorf("%s", string(command)+": command not found")
	}

	cmd := exec.Command(string(command), args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	clean := bytes.Trim(stdout.Bytes(), "\x00")
	output := string(clean)
	output = strings.TrimRight(output, "\n")

	if err != nil {
		result := string(stderr.String())
		result = strings.TrimRight(result, "\n")
		return output, fmt.Errorf("%s", result)
	}
	if len(stderr.Bytes()) > 0 {
		return "", fmt.Errorf("%s", stderr.String())
	}

	return output, nil
}

func (shell Shell) handlePwdCommand(args []string) (string, error) {
	return shell.directory, nil
}

func (shell *Shell) handleCdCommand(args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("expecting only 1 argument")
	} else if len(args) == 0 {
		return "", fmt.Errorf("missing argument")
	}

	goToPath := args[0]

	if goToPath == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "cd: error getting home directory", nil
		}
		err = os.Chdir(home)
		if err != nil {
			return "", err
		}
	} else {
		err := os.Chdir(goToPath)

		if err != nil {
			return "", fmt.Errorf("%s", "cd: "+goToPath+": No such file or directory")
		}
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	shell.directory = currentDir

	return "", nil
}

func filterParams(args []string) []string {
	newArgs := []string{}
	for _, item := range args {
		if item != " " {
			newArgs = append(newArgs, item)
		}
	}

	return newArgs
}

type CommandSpec struct {
	Name         Command
	NeedsRawArgs bool
	Handler      func(shell *Shell, args []string) (string, error)
}

var commands = map[Command]CommandSpec{
	EchoCommand: {EchoCommand, true, (*Shell).handleEchoCommand},
	TypeCommand: {TypeCommand, false, (*Shell).handleTypeCommand},
	PwdCommand:  {PwdCommand, false, (*Shell).handlePwdCommand},
	CdCommand:   {CdCommand, false, (*Shell).handleCdCommand},
}

func (shell *Shell) handleCommand(command Command, rawArgs []string) (string, error) {
	if spec, ok := commands[command]; ok {
		if spec.NeedsRawArgs {
			return spec.Handler(shell, rawArgs)
		}
		return spec.Handler(shell, filterParams(rawArgs))
	}

	return shell.handleExternalCommand(command, filterParams(rawArgs))
}

func (shell *Shell) redirect(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) == 1 {
		return "", fmt.Errorf("missing redirection dest")
	}

	operator := args[0]

	if operator != ">" && operator != "1>" && operator != "2>" {
		return "", fmt.Errorf("not supported redirection")
	}

	outputFile := args[1]
	file, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("%s", fmt.Sprintf("err creating file, err: %v", err))
	}
	if operator == "2>" {
		shell.stderr = file
	} else {
		shell.stdout = file
	}
	return "", nil

}

func (shell *Shell) startCli() (bool, int) {
	scanner := bufio.NewScanner(shell.in)
	stdout := shell.stdout
	stderr := shell.stderr
	for {
		shell.stdout = stdout
		shell.stderr = stderr
		fmt.Fprint(os.Stdout, "$ ")

		ok := scanner.Scan()
		if !ok {
			return false, 0
		}
		text := scanner.Text()

		parser := NewParser(text)
		input, err := parser.ParseCommand()
		if err != nil {
			fmt.Println(err)
			return true, 1
		}
		command := Command(input.Command)

		if command == ExitCommand {
			code := 0

			if len(input.Arguments) > 0 {
				n, err := strconv.Atoi(input.Arguments[0])
				if err == nil {
					code = n
				} else {
					code = 1
				}
			}

			return true, code
		}
		_, err = shell.redirect(input.Redirection)

		if err != nil {
			fmt.Fprintln(shell.stderr, err)
		}

		output, err := shell.handleCommand(input.Command, input.Arguments)

		if err != nil {
			fmt.Fprintln(shell.stderr, err)
		}

		if output != "" {
			fmt.Fprintln(shell.stdout, output)
		}

	}
}

func main() {
	directory, err := os.Getwd()

	if err != nil {
		fmt.Println("get directory err", err)
		os.Exit(1)
	}

	shell := Shell{
		in:        os.Stdin,
		stdout:    os.Stdout,
		stderr:    os.Stdout,
		directory: directory,
	}

	defer func() {
		if file, ok := shell.stdout.(*os.File); ok {
			file.Close()
		}
	}()

	exitRequest, exitCode := shell.startCli()

	if exitRequest {
		os.Exit(exitCode)
	}
}
