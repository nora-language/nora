"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.NoraDebugHoverProvider = void 0;
exports.activate = activate;
exports.deactivate = deactivate;
const path = require("path");
const fs = require("fs");
const child_process_1 = require("child_process");
const util_1 = require("util");
const vscode_1 = require("vscode");
const node_1 = require("vscode-languageclient/node");
const execFileAsync = (0, util_1.promisify)(child_process_1.execFile);
let client;
let clientStarted;
function getNoraExecutable(context) {
    const config = vscode_1.workspace.getConfiguration('nora');
    const configuredPath = config.get('server.path', '');
    if (configuredPath && configuredPath.trim() !== '') {
        return configuredPath;
    }
    const bundledPath = path.join(context.extensionPath, process.platform === 'win32' ? 'nora.exe' : 'nora');
    if (fs.existsSync(bundledPath)) {
        if (process.platform !== 'win32') {
            try {
                const stat = fs.statSync(bundledPath);
                if ((stat.mode & 0o100) === 0) {
                    fs.chmodSync(bundledPath, 0o755);
                }
            }
            catch (err) {
                console.error(`Failed to set execute permissions on ${bundledPath}:`, err);
            }
        }
        return bundledPath;
    }
    return process.platform === 'win32' ? 'nora.exe' : 'nora';
}
async function formatDocumentViaLsp(document) {
    await clientStarted;
    const editorConfig = vscode_1.workspace.getConfiguration('editor', document.uri);
    const tabSize = editorConfig.get('tabSize', 4);
    const insertSpaces = editorConfig.get('insertSpaces', true);
    const edits = await client.sendRequest('textDocument/formatting', {
        textDocument: { uri: document.uri.toString() },
        options: { tabSize, insertSpaces },
    });
    return edits ?? undefined;
}
async function applyFormatEdits(document, edits) {
    if (!edits || edits.length === 0) {
        return false;
    }
    const we = new vscode_1.WorkspaceEdit();
    we.set(document.uri, edits);
    return vscode_1.workspace.applyEdit(we);
}
function activate(context) {
    const config = vscode_1.workspace.getConfiguration('nora');
    const serverExe = getNoraExecutable(context);
    const serverOptions = {
        command: serverExe,
        args: ['lsp'],
        options: { shell: false },
    };
    const outputChannel = vscode_1.window.createOutputChannel('Nora Language Server');
    outputChannel.show(true);
    const clientOptions = {
        outputChannel: outputChannel,
        traceOutputChannel: outputChannel,
        documentSelector: [{ scheme: 'file', language: 'nr' }],
        synchronize: {
            fileEvents: vscode_1.workspace.createFileSystemWatcher('**/*.nr'),
        },
    };
    client = new node_1.LanguageClient('noraLanguageServer', 'Nora Language Server', serverOptions, clientOptions);
    clientStarted = client.start();
    context.subscriptions.push(vscode_1.commands.registerCommand('nora.formatDocument', async () => {
        const editor = vscode_1.window.activeTextEditor;
        if (!editor || editor.document.languageId !== 'nr') {
            void vscode_1.window.showWarningMessage('Open a .nr file to format.');
            return;
        }
        const edits = await formatDocumentViaLsp(editor.document);
        if (!(await applyFormatEdits(editor.document, edits))) {
            void vscode_1.window.showWarningMessage('Format failed. Fix parse errors in the file and try again.');
        }
    }));
    context.subscriptions.push(vscode_1.commands.registerCommand('nora.formatWorkspace', async () => {
        const folder = vscode_1.workspace.workspaceFolders?.[0];
        if (!folder) {
            void vscode_1.window.showWarningMessage('Open a folder workspace to format all .nr files.');
            return;
        }
        try {
            const { stdout, stderr } = await execFileAsync(serverExe, ['fmt', '-w'], { cwd: folder.uri.fsPath, maxBuffer: 10 * 1024 * 1024 });
            const msg = (stdout || stderr || '').trim();
            if (msg) {
                void vscode_1.window.showInformationMessage(msg);
            }
            else {
                void vscode_1.window.showInformationMessage('Workspace format complete.');
            }
        }
        catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            void vscode_1.window.showErrorMessage(`nora fmt failed: ${message}`);
        }
    }));
    // Premium Concurrency Debugger (DAP Tracker)
    context.subscriptions.push(vscode_1.debug.registerDebugAdapterTrackerFactory('*', {
        createDebugAdapterTracker(session) {
            return {
                onWillReceiveMessage: m => {
                    // Messages from VS Code TO the debug adapter
                    if (m.type === 'request' && m.command === 'stepIn') {
                        const editor = vscode_1.window.activeTextEditor;
                        if (editor && editor.document.languageId === 'nr') {
                            const line = editor.document.lineAt(editor.selection.active.line).text;
                            if (line.includes('spawn') || line.includes('parallel')) {
                                // Intercept stepIn and turn it into next (Step Over)
                                m.command = 'next';
                                // Set the global stepping flag in the C runtime
                                session.customRequest('evaluate', {
                                    expression: '__nora_stepping_into_spawn = 1',
                                    context: 'hover'
                                }).then(() => { }, () => { });
                            }
                        }
                    }
                    if (m.type === 'request' && m.command === 'continue') {
                        // Release scheduler locking on continue
                        session.customRequest('evaluate', {
                            expression: '__nora_debug_locked_fiber = 0',
                            context: 'hover'
                        }).then(() => { }, () => { });
                    }
                },
                onDidSendMessage: m => {
                    // Messages from Debug Adapter TO VS Code
                    if (m.type === 'event' && m.event === 'stopped') {
                        // Automatically lock the scheduler to the currently active fiber when stopped at any breakpoint/step
                        session.customRequest('evaluate', {
                            expression: '__nora_debug_locked_fiber = nr_fiber_current()',
                            context: 'hover'
                        }).then(() => { }, () => { });
                    }
                }
            };
        }
    }));
    if (config.get('format.onSave', false)) {
        context.subscriptions.push(vscode_1.workspace.onWillSaveTextDocument(async (event) => {
            if (event.document.languageId !== 'nr') {
                return;
            }
            const edits = await formatDocumentViaLsp(event.document);
            if (edits && edits.length > 0) {
                const we = new vscode_1.WorkspaceEdit();
                we.set(event.document.uri, edits);
                event.waitUntil(vscode_1.workspace.applyEdit(we));
            }
        }));
    }
    // Register debug hover provider to show both value and pointer address for heap allocated variables
    context.subscriptions.push(vscode_1.languages.registerHoverProvider({ scheme: 'file', language: 'nr' }, new NoraDebugHoverProvider()));
}
class NoraDebugHoverProvider {
    async provideHover(document, position, token) {
        const session = vscode_1.debug.activeDebugSession;
        if (!session) {
            return undefined;
        }
        const range = document.getWordRangeAtPosition(position);
        if (!range) {
            return undefined;
        }
        const word = document.getText(range);
        // Simple regex to check for valid variable name/expression
        if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(word)) {
            return undefined;
        }
        // Get active stack frame ID if available to evaluate in the correct scope
        let frameId = undefined;
        const debugAny = vscode_1.debug;
        if (debugAny.activeStackItem) {
            const item = debugAny.activeStackItem;
            if (item && typeof item.frameId === 'number') {
                frameId = item.frameId;
            }
        }
        try {
            // 1. Evaluate the variable as-is (pointer address if it is a heap-allocated variable)
            const evalResult = await session.customRequest('evaluate', {
                expression: word,
                frameId: frameId,
                context: 'hover'
            });
            if (!evalResult || !evalResult.result) {
                return undefined;
            }
            const addressOrValue = evalResult.result.trim();
            // 2. Try evaluating it as a dereferenced pointer (*word)
            try {
                const derefResult = await session.customRequest('evaluate', {
                    expression: `*${word}`,
                    frameId: frameId,
                    context: 'hover'
                });
                if (derefResult && derefResult.result) {
                    const derefValue = derefResult.result.trim();
                    // If the dereferenced result succeeded and is different from the original evaluation (e.g. is indeed a pointer)
                    if (addressOrValue !== derefValue) {
                        const md = new vscode_1.MarkdownString();
                        md.isTrusted = true;
                        md.appendMarkdown(`### Nora Debugger Hover\n\n`);
                        md.appendMarkdown(`* **Value**: \`${derefValue}\`\n`);
                        md.appendMarkdown(`* **Address**: \`${addressOrValue}\`\n`);
                        return new vscode_1.Hover(md, range);
                    }
                }
            }
            catch {
                // Dereferencing failed - word is not a pointer. Fall back to default hover.
                return undefined;
            }
        }
        catch {
            // Evaluation failed
            return undefined;
        }
        return undefined;
    }
}
exports.NoraDebugHoverProvider = NoraDebugHoverProvider;
function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
//# sourceMappingURL=extension.js.map