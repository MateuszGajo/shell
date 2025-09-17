package main

type Parser struct {
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

func (p Parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.input[p.i]
}

func (p *Parser) next() byte {
	p.i += 1
	if p.eof() {
		return 0
	}
	return p.input[p.i]
}

func (p Parser) eof() bool {
	return len(p.input) == p.i
}

func newParser(input string) Parser {
	return Parser{
		input: input,
		i:     0,
	}
}

func (p *Parser) nextToken() Token {
	var token Token
	switch p.peek() {
	case '\'':
		content := p.readSingleQuote()
		token = NewToken(STRING, content)
		p.next()
	case ' ':
		token = NewToken(SPACE, " ")
		p.next()
	case 0:
		token = NewToken(EOF, "")
	default:
		content := p.readLiteral()
		token = NewToken(STRING, content)
	}

	return token
}

func (p *Parser) readLiteral() string {
	start := p.i
	for isLiteral(p.next()) {
	}
	return p.input[start:p.i]
}

func isLiteral(b byte) bool {
	if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') {
		return true
	}

	return false
}

func (p *Parser) readSingleQuote() string {
	p.next()
	start := p.i
	for {
		aa := p.next()
		if aa == '\'' || aa == 0 {
			break
		}
	}
	return p.input[start:p.i]
}

func parseInput(input string) []string {
	result := []string{}

	parser := newParser(input)
	tokens := []Token{}

	token := parser.nextToken()
	for token.tokenType != EOF {
		tokens = append(tokens, token)
		token = parser.nextToken()
	}

	for _, token := range tokens {
		if token.tokenType == STRING {
			result = append(result, token.literal)
		}
	}

	return result
}
