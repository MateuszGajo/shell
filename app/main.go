package main

import (
	"bufio"
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
	out       io.Writer
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

func (shell Shell) handleEchoCommand(args []string) error {
	fmt.Fprintln(shell.out, strings.Join(args, ""))
	return nil
}

func (shell Shell) handleTypeCommand(args []string) error {

	if len(args) > 1 {
		return fmt.Errorf("expected only one argument")
	}

	if isBuiltinCommand(Command(args[0])) {
		fmt.Fprintln(shell.out, args[0]+" is a shell builtin")
	} else if ok, path := findFile(os.Getenv("PATH"), args[0]); ok {
		fmt.Fprintln(shell.out, args[0]+" is "+path)
	} else {
		fmt.Fprintln(shell.out, args[0]+": not found")
	}

	return nil
}

func (shell Shell) handleExternalCommand(command Command, args []string) error {
	ok, _ := findFile(os.Getenv("PATH"), string(command))

	if !ok {
		fmt.Fprintln(shell.out, command+": command not found")
		return nil
	}

	cmd := exec.Command(string(command), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(shell.out, err)
	}

	return nil
}

func (shell Shell) handlePwdCommand(args []string) error {
	fmt.Fprintln(shell.out, shell.directory)
	return nil
}

func (shell *Shell) handleCdCommand(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("expecting only 1 argument")
	} else if len(args) == 0 {
		return fmt.Errorf("missing argument")
	}

	goToPath := args[0]

	if goToPath == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(shell.out, "cd: error getting home directory")
		}
		err = os.Chdir(home)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cd: %v\n", err)
		}
	} else {
		err := os.Chdir(goToPath)

		if err != nil {
			fmt.Fprintln(shell.out, "cd: "+goToPath+": No such file or directory")
		}
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	shell.directory = currentDir

	return nil
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
	Handler      func(shell *Shell, args []string) error
}

var commands = map[Command]CommandSpec{
	EchoCommand: {EchoCommand, true, (*Shell).handleEchoCommand},
	TypeCommand: {TypeCommand, false, (*Shell).handleTypeCommand},
	PwdCommand:  {PwdCommand, false, (*Shell).handlePwdCommand},
	CdCommand:   {CdCommand, false, (*Shell).handleCdCommand},
}

func (shell *Shell) handleCommand(command Command, rawArgs []string) error {
	if spec, ok := commands[command]; ok {
		if spec.NeedsRawArgs {
			return spec.Handler(shell, rawArgs)
		}
		return spec.Handler(shell, filterParams(rawArgs))
	}

	return shell.handleExternalCommand(command, filterParams(rawArgs))

}

func (shell *Shell) startCli() (bool, int) {
	scanner := bufio.NewScanner(shell.in)
	for {
		fmt.Fprint(shell.out, "$ ")

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
					fmt.Fprintln(shell.out, "exit: invalid argument")
					code = 1
				}
			}

			return true, code
		}

		err = shell.handleCommand(input.Command, input.Arguments)

		if err != nil {
			fmt.Printf("handle command err %v", err)
			return true, 0
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
		out:       os.Stdout,
		directory: directory,
	}
	exitRequest, exitCode := shell.startCli()

	if exitRequest {
		os.Exit(exitCode)
	}
}
