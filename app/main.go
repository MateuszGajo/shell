package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Command string

const (
	EchoCommand Command = "echo"
	CDCommand   Command = "cd"
	ExitCommand Command = "exit"
	TypeCommand Command = "type"
)

var allowedCommands = []Command{EchoCommand, CDCommand, ExitCommand, TypeCommand}

type Shell struct {
	in  io.Reader
	out io.Writer
}

func isValidCommand(command Command) bool {
	for _, item := range allowedCommands {
		if item == command {
			return true
		}
	}
	return false
}

// Directories are splites by :
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
	fmt.Fprintln(shell.out, strings.Join(args, " "))
	return nil
}

func (shell Shell) handleTypeCommand(args []string) error {

	if len(args) > 1 {
		return fmt.Errorf("expected only one argument")
	}

	if isValidCommand(Command(args[0])) {
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
		return fmt.Errorf("couldn't find file")
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

func (shell Shell) handleCommand(command Command, args []string) error {

	switch command {
	case EchoCommand:
		return shell.handleEchoCommand(args)
	case TypeCommand:
		return shell.handleTypeCommand(args)
	default:
		err := shell.handleExternalCommand(command, args)
		if err != nil {
			fmt.Fprintln(shell.out, command+": command not found")
		}
	}

	return nil
}

func (shell Shell) startCli() (bool, int) {
	for {
		fmt.Fprint(shell.out, "$ ")

		text, err := bufio.NewReader(shell.in).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return false, 0
			}
			log.Fatal("Problem reading command", err)
		}
		split := strings.Split(text[:len(text)-1], " ")
		command := Command(split[0])

		if command == ExitCommand {
			code := 0

			if len(split) > 0 {
				n, err := strconv.Atoi(split[1])
				if err == nil {
					code = n
				} else {
					fmt.Fprintln(shell.out, "exit: invalid argument")
					code = 1
				}
			}

			return true, code
		}

		err = shell.handleCommand(command, split[1:])

		if err != nil {
			fmt.Printf("handle command err %v", err)
			return true, 0
		}

	}
}

func main() {
	shell := Shell{
		in:  os.Stdin,
		out: os.Stdout,
	}
	exitRequest, exitCode := shell.startCli()

	if exitRequest {
		os.Exit(exitCode)
	}
}
