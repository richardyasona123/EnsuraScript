import * as vscode from 'vscode';
import * as path from 'path';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext) {
    console.log('EnsuraScript extension activated');

    // Try to start LSP server
    startLanguageServer(context);

    // Fallback hover provider (used if LSP is not available)
    const hoverProvider = vscode.languages.registerHoverProvider('ens', {
        provideHover(document, position, token) {
            // If LSP is running, let it handle hovers
            if (client && client.isRunning()) {
                return null;
            }

            const range = document.getWordRangeAtPosition(position);
            const word = document.getText(range);

            const docs: Record<string, string> = {
                'ensure': 'Declares a guarantee that must be maintained.\n\n```ens\nensure <condition> [with <handler> <args>]\n```',
                'on': 'Opens a resource context block.\n\n```ens\non <resource-type> "<path>" { ... }\n```',
                'policy': 'Defines a reusable bundle of guarantees.\n\n```ens\npolicy <name>(<params>) { ... }\n```',
                'apply': 'Applies a policy to the current resource.\n\n```ens\napply <policy_name>(<args>)\n```',
                'with': 'Specifies a handler and its arguments.\n\n```ens\nensure encrypted with AES:256 key "env:KEY"\n```',
                'violation': 'Defines how to handle guarantee violations.\n\n```ens\nviolation { retry 3 times every 5s }\n```',
                'when': 'Guards a statement with a condition.\n\n```ens\nwhen environment == "production"\n```',
                'for': 'Iterates over a collection of resources.\n\n```ens\nfor each f in files("/etc/*.conf") { ... }\n```',
                'file': 'File resource type for filesystem paths.',
                'directory': 'Directory resource type for filesystem directories.',
                'http': 'HTTP resource type for web endpoints.',
                'service': 'Service resource type for system services.',
                'process': 'Process resource type for running processes.',
                'database': 'Database resource type for database connections.',
                'cron': 'Cron resource type for scheduled jobs.',
                'exists': 'Condition: Resource exists on the system.',
                'encrypted': 'Condition: Resource is encrypted. Implies exists.',
                'permissions': 'Condition: Resource has specific POSIX permissions.',
                'readable': 'Condition: Resource is readable.',
                'writable': 'Condition: Resource is writable.',
                'reachable': 'Condition: HTTP endpoint is reachable.',
                'running': 'Condition: Service or process is running.',
                'healthy': 'Condition: Service is healthy.',
                'tls': 'Condition: HTTP endpoint has valid TLS certificate.',
                'status_code': 'Condition: HTTP endpoint returns expected status.',
                'AES:256': 'Handler: AES-256-GCM encryption.\n\nArguments:\n- `key`: Encryption key reference (env:VAR or file:/path)',
                'posix': 'Handler: POSIX permission management.\n\nArguments:\n- `mode`: Octal permission mode (e.g., "0644")',
                'fs.native': 'Handler: Native filesystem operations.',
                'http.get': 'Handler: HTTP GET request handler.\n\nArguments:\n- `expected_status`: Expected HTTP status code',
            };

            if (docs[word]) {
                const markdown = new vscode.MarkdownString(docs[word]);
                markdown.isTrusted = true;
                return new vscode.Hover(markdown);
            }

            return null;
        }
    });

    context.subscriptions.push(hoverProvider);

    // Fallback document symbol provider
    const symbolProvider = vscode.languages.registerDocumentSymbolProvider('ens', {
        provideDocumentSymbols(document, token): vscode.ProviderResult<vscode.SymbolInformation[]> {
            // If LSP is running, let it handle symbols
            if (client && client.isRunning()) {
                return null;
            }

            const symbols: vscode.SymbolInformation[] = [];
            const text = document.getText();

            // Find policies
            const policyPattern = /policy\s+(\w+)/g;
            let match;
            while ((match = policyPattern.exec(text)) !== null) {
                const pos = document.positionAt(match.index);
                const range = new vscode.Range(pos, pos);
                symbols.push(new vscode.SymbolInformation(
                    match[1],
                    vscode.SymbolKind.Function,
                    '',
                    new vscode.Location(document.uri, range)
                ));
            }

            // Find resource blocks
            const resourcePattern = /on\s+(file|directory|http|service|process|database|cron)\s+"([^"]+)"/g;
            while ((match = resourcePattern.exec(text)) !== null) {
                const pos = document.positionAt(match.index);
                const range = new vscode.Range(pos, pos);
                symbols.push(new vscode.SymbolInformation(
                    `${match[1]}: ${match[2]}`,
                    vscode.SymbolKind.Object,
                    '',
                    new vscode.Location(document.uri, range)
                ));
            }

            return symbols;
        }
    });

    context.subscriptions.push(symbolProvider);
}

async function startLanguageServer(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('ensurascript');
    let serverPath = config.get<string>('lspPath');

    if (!serverPath) {
        // Try to find ensura-lsp in common locations
        const possiblePaths = [
            'ensura-lsp', // In PATH
            path.join(context.extensionPath, '..', '..', 'bin', 'ensura-lsp'),
        ];

        for (const p of possiblePaths) {
            try {
                const { execSync } = await import('child_process');
                execSync(`which ${p} 2>/dev/null || where ${p} 2>nul`, { stdio: 'pipe' });
                serverPath = p;
                break;
            } catch {
                // Continue to next path
            }
        }
    }

    if (!serverPath) {
        console.log('ensura-lsp not found, using fallback providers');
        return;
    }

    const serverOptions: ServerOptions = {
        command: serverPath,
        args: [],
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'ens' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.ens')
        }
    };

    client = new LanguageClient(
        'ensurascript',
        'EnsuraScript Language Server',
        serverOptions,
        clientOptions
    );

    try {
        await client.start();
        console.log('EnsuraScript LSP started');
    } catch (err) {
        console.log('Failed to start LSP:', err);
        client = undefined;
    }
}

export async function deactivate() {
    if (client) {
        await client.stop();
    }
}
