# Editor Support

EnsuraScript provides syntax highlighting and language support for popular editors.

## Shared Resources

The `shared/` directory contains reusable resources for editor integrations:

- `ensurascript.tmLanguage.json` - TextMate grammar for syntax highlighting

## VS Code

Full-featured VS Code extension with:

- Syntax highlighting
- Code snippets
- Hover documentation
- Document outline
- Language Server Protocol (LSP) support

### Installation

```bash
cd editors/vscode
npm install
npm run compile
npm run package
code --install-extension ensurascript-0.1.0.vsix
```

### LSP Configuration

To enable advanced features, install the LSP server and configure the path:

```bash
# Build the LSP server
go build -o ~/bin/ensura-lsp ./cmd/ensura-lsp

# In VS Code settings:
# "ensurascript.lspPath": "~/bin/ensura-lsp"
```

## Language Server

The `ensura-lsp` binary provides:

- Real-time diagnostics (parse errors)
- Hover information for keywords and handlers
- Code completion
- Document symbols for navigation

### Building

```bash
go build -o bin/ensura-lsp ./cmd/ensura-lsp
```

### Integration with Other Editors

The LSP server uses standard JSON-RPC over stdio, compatible with any LSP-capable editor:

**Neovim (nvim-lspconfig):**

```lua
require('lspconfig.configs').ensurascript = {
  default_config = {
    cmd = { 'ensura-lsp' },
    filetypes = { 'ens' },
    root_dir = function(fname)
      return vim.fn.getcwd()
    end,
  },
}
require('lspconfig').ensurascript.setup{}
```

**Emacs (lsp-mode):**

```elisp
(lsp-register-client
 (make-lsp-client
  :new-connection (lsp-stdio-connection '("ensura-lsp"))
  :major-modes '(ens-mode)
  :server-id 'ensurascript))
```
