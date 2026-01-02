package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ensurascript/ensura/pkg/ast"
	"github.com/ensurascript/ensura/pkg/lexer"
	"github.com/ensurascript/ensura/pkg/parser"
)

// LSP message types
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InitializeParams struct {
	ProcessID             int                `json:"processId"`
	RootURI               string             `json:"rootUri"`
	Capabilities          ClientCapabilities `json:"capabilities"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
}

type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Hover      *HoverCapability      `json:"hover,omitempty"`
	Completion *CompletionCapability `json:"completion,omitempty"`
}

type HoverCapability struct {
	ContentFormat []string `json:"contentFormat,omitempty"`
}

type CompletionCapability struct {
	CompletionItem *CompletionItemCapability `json:"completionItem,omitempty"`
}

type CompletionItemCapability struct {
	SnippetSupport bool `json:"snippetSupport,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

type ServerCapabilities struct {
	TextDocumentSync           int                     `json:"textDocumentSync"`
	HoverProvider              bool                    `json:"hoverProvider"`
	CompletionProvider         *CompletionOptions      `json:"completionProvider,omitempty"`
	DefinitionProvider         bool                    `json:"definitionProvider"`
	DocumentSymbolProvider     bool                    `json:"documentSymbolProvider"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// Diagnostic severity
const (
	DiagnosticSeverityError       = 1
	DiagnosticSeverityWarning     = 2
	DiagnosticSeverityInformation = 3
	DiagnosticSeverityHint        = 4
)

// Completion item kinds
const (
	CompletionKindKeyword  = 14
	CompletionKindFunction = 3
	CompletionKindProperty = 10
)

// Symbol kinds
const (
	SymbolKindFunction = 12
	SymbolKindObject   = 19
)

// Server state
type Server struct {
	documents map[string]string
	mu        sync.RWMutex
	writer    io.Writer
	writeMu   sync.Mutex
}

func NewServer(w io.Writer) *Server {
	return &Server{
		documents: make(map[string]string),
		writer:    w,
	}
}

func (s *Server) handleMessage(msg Message) {
	switch msg.Method {
	case "initialize":
		var params InitializeParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			s.sendError(msg.ID, -32602, "Invalid params")
			return
		}
		result := InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:       1, // Full sync
				HoverProvider:          true,
				DefinitionProvider:     true,
				DocumentSymbolProvider: true,
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{" ", "\""},
				},
			},
			ServerInfo: &ServerInfo{
				Name:    "ensura-lsp",
				Version: "0.1.0",
			},
		}
		s.sendResult(msg.ID, result)

	case "initialized":
		// Client acknowledged initialization

	case "shutdown":
		s.sendResult(msg.ID, nil)

	case "exit":
		os.Exit(0)

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return
		}
		s.mu.Lock()
		s.documents[params.TextDocument.URI] = params.TextDocument.Text
		s.mu.Unlock()
		s.publishDiagnostics(params.TextDocument.URI)

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return
		}
		if len(params.ContentChanges) > 0 {
			s.mu.Lock()
			s.documents[params.TextDocument.URI] = params.ContentChanges[len(params.ContentChanges)-1].Text
			s.mu.Unlock()
			s.publishDiagnostics(params.TextDocument.URI)
		}

	case "textDocument/didClose":
		var params struct {
			TextDocument TextDocumentIdentifier `json:"textDocument"`
		}
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return
		}
		s.mu.Lock()
		delete(s.documents, params.TextDocument.URI)
		s.mu.Unlock()

	case "textDocument/hover":
		var params TextDocumentPositionParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			s.sendError(msg.ID, -32602, "Invalid params")
			return
		}
		hover := s.getHover(params)
		s.sendResult(msg.ID, hover)

	case "textDocument/completion":
		var params TextDocumentPositionParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			s.sendError(msg.ID, -32602, "Invalid params")
			return
		}
		completions := s.getCompletions(params)
		s.sendResult(msg.ID, completions)

	case "textDocument/documentSymbol":
		var params struct {
			TextDocument TextDocumentIdentifier `json:"textDocument"`
		}
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			s.sendError(msg.ID, -32602, "Invalid params")
			return
		}
		symbols := s.getDocumentSymbols(params.TextDocument.URI)
		s.sendResult(msg.ID, symbols)
	}
}

func (s *Server) publishDiagnostics(uri string) {
	s.mu.RLock()
	content, ok := s.documents[uri]
	s.mu.RUnlock()

	if !ok {
		return
	}

	diagnostics := []Diagnostic{}

	// Parse the document
	l := lexer.New(content)
	p := parser.New(l)
	_ = p.Parse()

	for _, err := range p.Errors() {
		// Try to extract line info from error message
		line := 0
		col := 0
		msg := err

		// Simple parsing of "line X" pattern
		if strings.Contains(err, "line ") {
			parts := strings.Split(err, "line ")
			if len(parts) > 1 {
				numPart := strings.Split(parts[1], ":")[0]
				numPart = strings.Split(numPart, " ")[0]
				if n, e := strconv.Atoi(numPart); e == nil {
					line = n - 1 // LSP uses 0-based lines
				}
			}
		}

		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 10},
			},
			Severity: DiagnosticSeverityError,
			Message:  msg,
			Source:   "ensura",
		})
	}

	params := PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	s.sendNotification("textDocument/publishDiagnostics", params)
}

func (s *Server) getHover(params TextDocumentPositionParams) *Hover {
	s.mu.RLock()
	content, ok := s.documents[params.TextDocument.URI]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	word := getWordAtPosition(content, params.Position)
	if word == "" {
		return nil
	}

	docs := map[string]string{
		"ensure":      "Declares a guarantee that must be maintained.\n\n```ens\nensure <condition> [with <handler> <args>]\n```",
		"on":          "Opens a resource context block.\n\n```ens\non <resource-type> \"<path>\" { ... }\n```",
		"policy":      "Defines a reusable bundle of guarantees.\n\n```ens\npolicy <name>(<params>) { ... }\n```",
		"apply":       "Applies a policy to the current resource.\n\n```ens\napply <policy_name>(<args>)\n```",
		"with":        "Specifies a handler and its arguments.",
		"violation":   "Defines how to handle guarantee violations.",
		"when":        "Guards a statement with a condition.",
		"for":         "Iterates over a collection of resources.",
		"file":        "File resource type for filesystem paths.",
		"directory":   "Directory resource type for filesystem directories.",
		"http":        "HTTP resource type for web endpoints.",
		"service":     "Service resource type for system services.",
		"process":     "Process resource type for running processes.",
		"database":    "Database resource type for database connections.",
		"cron":        "Cron resource type for scheduled jobs.",
		"exists":      "Condition: Resource exists on the system.",
		"encrypted":   "Condition: Resource is encrypted. Implies `exists`.",
		"permissions": "Condition: Resource has specific POSIX permissions.",
		"readable":    "Condition: Resource is readable.",
		"writable":    "Condition: Resource is writable.",
		"reachable":   "Condition: HTTP endpoint is reachable.",
		"running":     "Condition: Service or process is running.",
		"healthy":     "Condition: Service is healthy.",
		"tls":         "Condition: HTTP endpoint has valid TLS certificate.",
		"status_code": "Condition: HTTP endpoint returns expected status.",
		"AES:256":     "Handler: AES-256-GCM encryption.\n\nArguments:\n- `key`: Encryption key reference",
		"posix":       "Handler: POSIX permission management.\n\nArguments:\n- `mode`: Octal permission mode",
		"fs.native":   "Handler: Native filesystem operations.",
		"http.get":    "Handler: HTTP GET request handler.",
	}

	if doc, ok := docs[word]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: doc,
			},
		}
	}

	return nil
}

func (s *Server) getCompletions(params TextDocumentPositionParams) *CompletionList {
	items := []CompletionItem{
		{Label: "ensure", Kind: CompletionKindKeyword, Detail: "Declare a guarantee"},
		{Label: "on", Kind: CompletionKindKeyword, Detail: "Open resource context"},
		{Label: "policy", Kind: CompletionKindKeyword, Detail: "Define a policy"},
		{Label: "apply", Kind: CompletionKindKeyword, Detail: "Apply a policy"},
		{Label: "with", Kind: CompletionKindKeyword, Detail: "Specify handler"},
		{Label: "when", Kind: CompletionKindKeyword, Detail: "Guard condition"},
		{Label: "violation", Kind: CompletionKindKeyword, Detail: "Violation handler"},
		{Label: "for", Kind: CompletionKindKeyword, Detail: "Loop construct"},

		{Label: "file", Kind: CompletionKindProperty, Detail: "File resource"},
		{Label: "directory", Kind: CompletionKindProperty, Detail: "Directory resource"},
		{Label: "http", Kind: CompletionKindProperty, Detail: "HTTP resource"},
		{Label: "service", Kind: CompletionKindProperty, Detail: "Service resource"},
		{Label: "process", Kind: CompletionKindProperty, Detail: "Process resource"},
		{Label: "database", Kind: CompletionKindProperty, Detail: "Database resource"},
		{Label: "cron", Kind: CompletionKindProperty, Detail: "Cron resource"},

		{Label: "exists", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "encrypted", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "permissions", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "readable", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "writable", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "reachable", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "running", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "healthy", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "tls", Kind: CompletionKindFunction, Detail: "Condition"},
		{Label: "status_code", Kind: CompletionKindFunction, Detail: "Condition"},

		{Label: "AES:256", Kind: CompletionKindProperty, Detail: "Encryption handler"},
		{Label: "posix", Kind: CompletionKindProperty, Detail: "Permission handler"},
		{Label: "fs.native", Kind: CompletionKindProperty, Detail: "Filesystem handler"},
		{Label: "http.get", Kind: CompletionKindProperty, Detail: "HTTP handler"},
	}

	return &CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

func (s *Server) getDocumentSymbols(uri string) []DocumentSymbol {
	s.mu.RLock()
	content, ok := s.documents[uri]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	symbols := []DocumentSymbol{}

	l := lexer.New(content)
	p := parser.New(l)
	program := p.Parse()

	if program == nil {
		return symbols
	}

	for _, stmt := range program.Statements {
		switch st := stmt.(type) {
		case *ast.PolicyDecl:
			pos := st.Pos()
			symbols = append(symbols, DocumentSymbol{
				Name: st.Name,
				Kind: SymbolKindFunction,
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column + len(st.Name)},
				},
				SelectionRange: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column + len(st.Name)},
				},
			})
		case *ast.OnBlock:
			pos := st.Pos()
			name := fmt.Sprintf("%s: %s", st.Subject.ResourceType, st.Subject.Path)
			symbols = append(symbols, DocumentSymbol{
				Name: name,
				Kind: SymbolKindObject,
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column + len(name)},
				},
				SelectionRange: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column + len(name)},
				},
			})
		}
	}

	return symbols
}

func getWordAtPosition(content string, pos Position) string {
	lines := strings.Split(content, "\n")
	if pos.Line >= len(lines) {
		return ""
	}

	line := lines[pos.Line]
	if pos.Character >= len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == ':' || c == '.'
}

func (s *Server) sendResult(id *json.RawMessage, result interface{}) {
	msg := Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeMessage(msg)
}

func (s *Server) sendError(id *json.RawMessage, code int, message string) {
	msg := Message{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ResponseError{
			Code:    code,
			Message: message,
		},
	}
	s.writeMessage(msg)
}

func (s *Server) sendNotification(method string, params interface{}) {
	paramsJSON, _ := json.Marshal(params)
	msg := Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsJSON,
	}
	s.writeMessage(msg)
}

func (s *Server) writeMessage(msg Message) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	content, err := json.Marshal(msg)
	if err != nil {
		return
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	s.writer.Write([]byte(header))
	s.writer.Write(content)
}

func main() {
	server := NewServer(os.Stdout)
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read headers
		var contentLength int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				continue
			}

			line = strings.TrimSpace(line)
			if line == "" {
				break
			}

			if strings.HasPrefix(line, "Content-Length:") {
				fmt.Sscanf(line, "Content-Length: %d", &contentLength)
			}
		}

		if contentLength == 0 {
			continue
		}

		// Read content
		content := make([]byte, contentLength)
		_, err := io.ReadFull(reader, content)
		if err != nil {
			continue
		}

		var msg Message
		if err := json.Unmarshal(content, &msg); err != nil {
			continue
		}

		server.handleMessage(msg)
	}
}
