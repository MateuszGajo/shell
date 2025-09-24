package main

import (
	"fmt"
	"slices"
)

// Lets use descent parser, ll(1)

// Grammar
// command -> String spaces argument_list redirection_list
// argument_list -> String spaces argument_list | ε
// redirection_list -> redirection spaces redirection_list | ε
// redirection -> redirect_op spaces String
// spaces -> SPACE | ε
// redirect_op -> ">" | ">>" | "<" | "2>" | "&>" | "1>"

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
	STRING   = "STRRING"
	SPACE    = "SPACE"
	EOF      = "EOF"
	EPSILON  = "EPSILON"
	REDIRECT = "REDIRECT"
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

func (l Lexar) peekNext() byte {
	newIndex := l.i
	newIndex++
	if l.indexEof(newIndex) {
		return 0
	}

	return l.input[newIndex]
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

func (p Lexar) indexEof(index int) bool {
	return len(p.input) == index
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
	case '1', '2':
		start := p.i
		if p.peekNext() != '>' {
			result := p.readLiteral()
			token = NewToken(STRING, result)
			break
		}
		p.next()
		if p.next() == '>' {
			p.next()
		}
		token = NewToken(REDIRECT, string(p.input[start:p.i]))
	case '>':
		start := p.i
		if p.next() == '>' {
			p.next()
		}
		token = NewToken(REDIRECT, string(p.input[start:p.i]))
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

	var token Token
	for {
		token = p.lexar.nextToken()
		if token.tokenType != EPSILON {
			break
		}
	}
	p.peekToken = token

}

type ParsedInput struct {
	Command     Command
	Arguments   []string
	Redirection []string
}

func (p *Parser) parseSpaces() {
	if p.currentToken.tokenType == SPACE {
		p.nextToken() // consume SPACE
	}
}

// command -> String spaces argument_list redirection_list
func (p *Parser) ParseCommand() (ParsedInput, error) {

	if p.currentToken.tokenType != STRING {
		return ParsedInput{}, fmt.Errorf("expect command, got: %s", p.currentToken.tokenType)
	}

	command := ParsedInput{
		Command:     Command(p.currentToken.literal),
		Arguments:   []string{},
		Redirection: []string{},
	}

	p.nextToken()
	p.parseSpaces()

	command.Arguments = p.ParseArgumentList()
	command.Redirection = p.parseRedirectionList()

	return command, nil

}

// redirection_list -> redirection spaces redirection_list | ε
// redirection -> redirect_op spaces String
// spaces -> SPACE | ε
// redirect_op -> ">" | ">>" | "<" | "2>" | "&>" | "1>"
func (p *Parser) parseRedirectionList() []string {
	var list []string

	if p.currentToken.tokenType == EOF || p.currentToken.tokenType != REDIRECT {
		return list
	}

	list = append(list, p.currentToken.literal)
	p.nextToken()
	p.parseSpaces()
	if p.currentToken.tokenType != STRING {
		return list
	}
	list = append(list, p.currentToken.literal)
	p.nextToken()

	moreRedirection := p.parseRedirectionList()
	list = append(list, moreRedirection...)

	return list

}

// argument_list -> String spaces argument_list | ε
func (p *Parser) ParseArgumentList() []string {

	var args []string

	if p.currentToken.tokenType == EOF || p.currentToken.tokenType == REDIRECT {
		return args
	}

	args = append(args, p.currentToken.literal)
	p.nextToken()

	moreArgs := p.ParseArgumentList()

	args = append(args, moreArgs...)

	return args
}
