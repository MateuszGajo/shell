package main

import "fmt"

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
	case 0:
		token = NewToken(EOF, "")
	default:
		content := p.readConcatenatedString()
		token = NewToken(STRING, content)
	}

	return token
}

func (p *Lexar) readConcatenatedString() string {
	var result string

	for !p.eof() && p.peek() != ' ' && p.peek() != 0 {
		switch p.peek() {
		case '\'':
			result += p.readSingleQuote()
			p.next()
		case '"':
			result += p.readDoubleQuote()
			p.next()
		default:
			result += p.readLiteral()
		}
	}

	return result
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

func (p *Lexar) readDoubleQuote() string {
	char := p.next()
	start := p.i
	for {
		if char == '"' || char == 0 {
			break
		}
		char = p.next()

	}
	return p.input[start:p.i]
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

	for {
		p.peekToken = p.lexar.nextToken()
		if p.peekToken.tokenType != SPACE {
			break
		}
	}
}

type ParsedInput struct {
	Command   Command
	Arguments []string
}

func (p *Parser) ParseCommand() (ParsedInput, error) {
	//Grammar rule command -> String argument_list

	if p.currentToken.tokenType != STRING {
		return ParsedInput{}, fmt.Errorf("Expect command, got: %s", p.currentToken.tokenType)
	}

	command := ParsedInput{
		Command:   Command(p.currentToken.literal),
		Arguments: []string{},
	}

	p.nextToken()

	command.Arguments = p.ParseArgumentList()

	return command, nil

}

func (p *Parser) ParseArgumentList() []string {
	// Grammar rule: argument_list -> STRING argument_list | ε

	var args []string

	if p.currentToken.tokenType != STRING {
		return args
	}

	args = append(args, p.currentToken.literal)
	p.nextToken()

	moreArgs := p.ParseArgumentList()

	args = append(args, moreArgs...)

	return args
}

// func (p Parser) parseInput() []string {
// 	result := []string{}

// 	parser := newLexar(p.input)
// 	tokens := []Token{}

// 	token := parser.nextToken()
// 	for token.tokenType != EOF {
// 		tokens = append(tokens, token)
// 		token = parser.nextToken()
// 	}

// 	literal := ""
// 	for _, token := range tokens {
// 		if token.tokenType == STRING {
// 			literal += token.literal
// 			// result = append(result, token.literal)
// 		} else if token.tokenType == SPACE {
// 			result = append(result, literal)
// 			literal = ""
// 		}

// 	}
// 	if len(literal) > 0 {
// 		result = append(result, literal)
// 	}

// 	return result
// }
