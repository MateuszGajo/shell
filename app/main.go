package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
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
	} else {
		fmt.Fprintln(shell.out, args[0]+": not found")
	}

	return nil
}

func (shell Shell) handleCommand(command Command, args []string) error {
	var err error
	switch command {
	case EchoCommand:
		err = shell.handleEchoCommand(args)
	case TypeCommand:
		err = shell.handleTypeCommand(args)
	default:
		err = fmt.Errorf("Command handler not defined for: %v", command)
	}

	if err != nil {
		return fmt.Errorf("Command: %v, err:%v", command, err)
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

		if !isValidCommand(command) {
			fmt.Fprintln(shell.out, command+": command not found")
			continue
		}

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
