package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `resource file "secrets.db"
ensure exists on file "secrets.db"
ensure encrypted with AES:256 key "env:SECRET_KEY"
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{RESOURCE, "resource"},
		{FILE, "file"},
		{STRING, "secrets.db"},
		{ENSURE, "ensure"},
		{IDENT, "exists"},
		{ON, "on"},
		{FILE, "file"},
		{STRING, "secrets.db"},
		{ENSURE, "ensure"},
		{IDENT, "encrypted"},
		{WITH, "with"},
		{IDENT, "AES"},
		{COLON, ":"},
		{NUMBER, "256"},
		{KEY, "key"},
		{STRING, "env:SECRET_KEY"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"resource", RESOURCE},
		{"ensure", ENSURE},
		{"on", ON},
		{"with", WITH},
		{"requires", REQUIRES},
		{"after", AFTER},
		{"before", BEFORE},
		{"policy", POLICY},
		{"apply", APPLY},
		{"retry", RETRY},
		{"notify", NOTIFY},
		{"assume", ASSUME},
		{"when", WHEN},
		{"for", FOR},
		{"each", EACH},
		{"in", IN},
		{"invariant", INVARIANT},
		{"as", AS},
		{"key", KEY},
		{"mode", MODE},
		{"file", FILE},
		{"directory", DIRECTORY},
		{"http", HTTP},
		{"database", DATABASE},
		{"unknown_identifier", IDENT},
	}

	for _, tt := range tests {
		result := LookupIdent(tt.input)
		if result != tt.expected {
			t.Errorf("LookupIdent(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestComments(t *testing.T) {
	input := `# This is a comment
resource file "test.txt"
# Another comment
ensure exists`

	l := New(input)
	tokens := l.Tokenize() // Tokenize skips comments

	// Should have: resource, file, string, ensure, ident, EOF
	if len(tokens) != 6 {
		t.Errorf("Expected 6 tokens (comments stripped), got %d", len(tokens))
	}

	// TokenizeAll includes comments
	l2 := New(input)
	allTokens := l2.TokenizeAll()

	commentCount := 0
	for _, tok := range allTokens {
		if tok.Type == COMMENT {
			commentCount++
		}
	}

	if commentCount != 2 {
		t.Errorf("Expected 2 comments, got %d", commentCount)
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"path/to/file"`, "path/to/file"},
		{`"env:SECRET_KEY"`, "env:SECRET_KEY"},
		{`"https://example.com"`, "https://example.com"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != STRING {
			t.Errorf("Expected STRING, got %v", tok.Type)
		}

		if tok.Literal != tt.expected {
			t.Errorf("Expected %q, got %q", tt.expected, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `environment == "prod"
status != "failed"`

	l := New(input)

	// environment
	tok := l.NextToken()
	if tok.Type != ENVIRONMENT {
		t.Errorf("Expected ENVIRONMENT, got %v", tok.Type)
	}

	// ==
	tok = l.NextToken()
	if tok.Type != EQUALS {
		t.Errorf("Expected EQUALS, got %v", tok.Type)
	}

	// "prod"
	tok = l.NextToken()
	if tok.Type != STRING {
		t.Errorf("Expected STRING, got %v", tok.Type)
	}

	// status
	tok = l.NextToken()
	if tok.Type != IDENT {
		t.Errorf("Expected IDENT, got %v", tok.Type)
	}

	// !=
	tok = l.NextToken()
	if tok.Type != NOTEQUALS {
		t.Errorf("Expected NOTEQUALS, got %v", tok.Type)
	}
}

func TestPosition(t *testing.T) {
	input := `resource file "test.txt"
ensure exists`

	l := NewWithFilename(input, "test.ens")

	tok := l.NextToken()
	if tok.Pos.Line != 1 {
		t.Errorf("Expected line 1, got %d", tok.Pos.Line)
	}
	if tok.Pos.Filename != "test.ens" {
		t.Errorf("Expected filename 'test.ens', got %s", tok.Pos.Filename)
	}

	// Skip to next line
	l.NextToken() // file
	l.NextToken() // "test.txt"
	tok = l.NextToken() // ensure

	if tok.Pos.Line != 2 {
		t.Errorf("Expected line 2, got %d", tok.Pos.Line)
	}
}
