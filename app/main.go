package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
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

func displayFilesFromDir(directories string) []string {
	directoriesSplit := strings.Split(directories, ":")
	allFiles := []string{}

	for _, item := range directoriesSplit {
		entries, err := os.ReadDir(item)

		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				allFiles = append(allFiles, entry.Name())
			}
		}

	}
	return allFiles
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

func (shell *Shell) handleExternalCommand(command Command, args []string) (*exec.Cmd, error) {
	ok, _ := findFile(os.Getenv("PATH"), string(command))

	if !ok {
		return nil, fmt.Errorf("%s", string(command)+": command not found")
	}

	cmd := exec.Command(string(command), args...)

	return cmd, nil
}

func (shell *Shell) pipelineFunctionWrapper(r *os.File, w *os.File, args []string, fn func(shell *Shell, args []string) (string, error)) {
	go func() {
		output, err := fn(shell, args)
		w.Write([]byte(output + "\n"))

		if err != nil {
			panic("err")
		}
	}()
}

type Pipes struct {
	write *os.File
	read  *os.File
}

func (shell *Shell) pipeline(commands []ParsedCommand) error {
	var rPrev *os.File

	var cmds []*exec.Cmd

	pipesIO := []Pipes{}

	for i := 0; i < len(commands)-1; i++ {
		r, w, err := os.Pipe()

		if err != nil {
			panic(err)
		}
		pipesIO = append(pipesIO, Pipes{write: w, read: rPrev})

		rPrev = r
	}

	var writerStdOut io.Writer = shell.stdout

	stdOutFile, ok := writerStdOut.(*os.File)
	if !ok {
		return fmt.Errorf("couldnt convert shell stdout to file")
	}
	pipesIO = append(pipesIO, Pipes{write: stdOutFile, read: rPrev})

	for index, comamnd := range commands {
		handlerFunc := shell.getHandleCommandRaw(comamnd.Command)
		args := filterParams(comamnd.Arguments)

		if handlerFunc.SimpleHandler != nil {
			shell.pipelineFunctionWrapper(pipesIO[index].read, pipesIO[index].write, args, handlerFunc.SimpleHandler)
		} else if handlerFunc.CommandHandler != nil {
			exec, err := handlerFunc.CommandHandler(shell, comamnd.Command, args)
			if err != nil {
				return fmt.Errorf("eror while getting cmd setup, %v", err)
			}
			if index > 0 {

				exec.Stdin = pipesIO[index].read
			}
			exec.Stdout = pipesIO[index].write
			exec.Stderr = shell.stderr

			err = exec.Start()

			if err != nil {
				return fmt.Errorf("eror while starting cmd, %v", err)
			}
			cmds = append(cmds, exec)
		}

	}

	for i := 0; i < len(pipesIO)-1; i++ {
		pipesIO[i].write.Close()
	}
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return err
		}
	}

	for _, r := range pipesIO {
		r.read.Close()
	}

	return nil

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

type CommandSpecResponse struct {
	SimpleHandler  func(shell *Shell, args []string) (string, error)
	CommandHandler func(shell *Shell, command Command, args []string) (*exec.Cmd, error)
}

var commands = map[Command]CommandSpec{
	EchoCommand: {EchoCommand, true, (*Shell).handleEchoCommand},
	TypeCommand: {TypeCommand, false, (*Shell).handleTypeCommand},
	PwdCommand:  {PwdCommand, false, (*Shell).handlePwdCommand},
	CdCommand:   {CdCommand, false, (*Shell).handleCdCommand},
}

func (shell *Shell) getHandleCommandRaw(command Command) CommandSpecResponse {
	data, ok := commands[command]
	if ok {
		return CommandSpecResponse{
			SimpleHandler:  data.Handler,
			CommandHandler: nil,
		}
	}

	return CommandSpecResponse{
		SimpleHandler:  nil,
		CommandHandler: (*Shell).handleExternalCommand,
	}
}

func (shell *Shell) filterArgs(command Command, args []string) []string {
	if spec, ok := commands[command]; !(ok && spec.NeedsRawArgs) {
		args = filterParams(args)

	}

	return args
}

func (shell *Shell) handleCommand(command Command, rawArgs []string) (string, error) {

	handlerFunc := shell.getHandleCommandRaw(command)
	args := shell.filterArgs(command, rawArgs)
	if handlerFunc.SimpleHandler != nil {
		return handlerFunc.SimpleHandler(shell, args)
	} else if handlerFunc.CommandHandler == nil {
		panic("Never should happend that command handler is nill when simpler handler inill too")
	}

	cmd, err := handlerFunc.CommandHandler(shell, command, args)

	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
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

func (shell *Shell) redirect(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	if len(args) == 1 {
		return "", fmt.Errorf("missing redirection dest")
	}

	operator := args[0]

	if operator != ">" && operator != "1>" && operator != "2>" && operator != ">>" && operator != "1>>" && operator != "2>>" {
		return "", fmt.Errorf("not supported redirection")
	}

	append := strings.Contains(operator, ">>")

	outputFile := args[1]
	var file *os.File
	var err error

	if append {
		file, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(outputFile)
	}

	if err != nil {
		return "", fmt.Errorf("%s", fmt.Sprintf("err creating file, err: %v", err))
	}
	if operator == "2>" || operator == "2>>" {
		shell.stderr = file
	} else {
		shell.stdout = file
	}
	return "", nil

}

func (shell *Shell) startCli() (bool, int) {
	stdout := shell.stdout
	stderr := shell.stderr
	stdin := shell.in

	l, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		Stdin:        io.NopCloser(shell.in),
		AutoComplete: &AutoComplete{},
	})
	if err != nil {
		return true, 0
	}
	for {
		shell.stdout = stdout
		shell.stderr = stderr
		shell.in = stdin
		fmt.Fprint(os.Stdout, "$ ")

		raw, err := l.Readline()
		if err != nil {
			return false, 0
		}

		parser := NewParser(raw)
		commands, err := parser.parsePipe()
		if err != nil {
			fmt.Println(err)
			return true, 1
		}
		if len(commands) > 1 {
			shell.pipeline(commands)

			continue
		}
		if len(commands) == 0 {
			fmt.Fprintf(shell.stdout, "No command to do")
			return true, 1
		}
		input := commands[0]

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
