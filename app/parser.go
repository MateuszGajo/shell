package main

import (
	"fmt"
	"slices"
)

// Lets use recusrive descent parser

// Grammar
// command -> String argument_list
// argument_list -> String argument_list | ε

type Lexar struct {
	i     int
	input string
}

type TokenType string

type Token struct {
	tokenType TokenType
	literal   string
}

const (
	STRING = "STRRING"
	SPACE  = "SPACE"
	EOF    = "EOF"
)

func NewToken(tokenType TokenType, literal string) Token {
	return Token{
		tokenType: tokenType,
		literal:   literal,
	}
}

func (p Lexar) peek() byte {
	if p.eof() {
		return 0
	}
	return p.input[p.i]
}

func (p *Lexar) next() byte {
	p.i += 1
	if p.eof() {
		return 0
	}
	return p.input[p.i]
}

func (p Lexar) eof() bool {
	return len(p.input) == p.i
}

func newLexar(input string) Lexar {
	return Lexar{
		input: input,
		i:     0,
	}
}

func (p *Lexar) nextToken() Token {
	var token Token
	switch p.peek() {
	case ' ':
		p.readAllSpace()
		token = NewToken(SPACE, " ")
	case '\\':
		b := p.readEscapedByte()
		token = NewToken(STRING, string(b))
	case '\'':
		result := p.readSingleQuote()
		p.next()
		token = NewToken(STRING, result)
	case '"':
		result := p.readDoubleQuote()
		p.next()
		token = NewToken(STRING, result)
	case 0:
		token = NewToken(EOF, "")
	default:
		result := p.readLiteral()
		token = NewToken(STRING, result)
	}

	return token
}

func (l *Lexar) readEscapedByte() byte {
	ret := l.next()
	l.next()

	return ret
}

func (l *Lexar) readAllSpace() {
	for l.next() == ' ' {
	}
}

func (p *Lexar) readLiteral() string {
	start := p.i
	for isLiteral(p.next()) {
	}
	return p.input[start:p.i]
}

func isLiteral(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '/' || b == '_' || b == '-' || b == '.'

}

func (p *Lexar) readSingleQuote() string {
	char := p.next()
	start := p.i
	for {
		if char == '\'' || char == 0 {
			break
		}
		char = p.next()

	}
	return p.input[start:p.i]
}

var doubleQuoteEscapables = []byte{'"', '\\', '$', '`', '\n'}

func (p *Lexar) readDoubleQuote() string {
	char := p.next()
	res := ""
	for {
		if char == '"' || char == 0 {
			break
		}
		if char == '\\' {
			currentChar := char
			char = p.next()

			if !slices.Contains(doubleQuoteEscapables, char) {
				res += string(currentChar)
			}
		}

		res += string(char)
		char = p.next()

	}
	// res += string(char)
	return res
}

type Parser struct {
	lexar        Lexar
	currentToken Token
	peekToken    Token
}

func NewParser(input string) Parser {
	p := Parser{
		lexar: newLexar(input),
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken

	p.peekToken = p.lexar.nextToken()

}

type ParsedInput struct {
	Command   Command
	Arguments []string
}

func (p *Parser) ParseCommand() (ParsedInput, error) {
	//Grammar rule command -> String argument_list

	if p.currentToken.tokenType != STRING {
		return ParsedInput{}, fmt.Errorf("expect command, got: %s", p.currentToken.tokenType)
	}

	command := ParsedInput{
		Command:   Command(p.currentToken.literal),
		Arguments: []string{},
	}

	p.nextToken()

	p.nextToken()

	command.Arguments = p.ParseArgumentList()

	return command, nil

}

func (p *Parser) ParseArgumentList() []string {
	// Grammar rule: argument_list -> STRING argument_list | ε

	var args []string

	if p.currentToken.tokenType == EOF {
		return args
	}

	args = append(args, p.currentToken.literal)
	p.nextToken()

	moreArgs := p.ParseArgumentList()

	args = append(args, moreArgs...)

	return args
}
