import * as path from 'path';
import * as fs from 'fs';
import { execFile } from 'child_process';
import { promisify } from 'util';
import {
	workspace,
	ExtensionContext,
	commands,
	window,
	WorkspaceEdit,
	TextDocument,
	TextEdit,
	languages,
	Hover,
	HoverProvider,
	MarkdownString,
	Position,
	Range,
	CancellationToken,
	debug,
	DebugSession,
	FunctionBreakpoint,
} from 'vscode';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	Executable,
} from 'vscode-languageclient/node';

const execFileAsync = promisify(execFile);

let client: LanguageClient;
let clientStarted: Promise<void>;

function getNoraExecutable(context: ExtensionContext): string {
	const config = workspace.getConfiguration('nora');
	const configuredPath = config.get<string>('server.path', '');
	if (configuredPath && configuredPath.trim() !== '') {
		return configuredPath;
	}
	const bundledPath = path.join(
		context.extensionPath,
		process.platform === 'win32' ? 'nora.exe' : 'nora'
	);
	if (fs.existsSync(bundledPath)) {
		if (process.platform !== 'win32') {
			try {
				const stat = fs.statSync(bundledPath);
				if ((stat.mode & 0o100) === 0) {
					fs.chmodSync(bundledPath, 0o755);
				}
			} catch (err) {
				console.error(`Failed to set execute permissions on ${bundledPath}:`, err);
			}
		}
		return bundledPath;
	}
	return process.platform === 'win32' ? 'nora.exe' : 'nora';
}


async function formatDocumentViaLsp(
	document: TextDocument
): Promise<TextEdit[] | undefined> {
	await clientStarted;
	const editorConfig = workspace.getConfiguration('editor', document.uri);
	const tabSize = editorConfig.get<number>('tabSize', 4);
	const insertSpaces = editorConfig.get<boolean>('insertSpaces', true);

	const edits = await client.sendRequest<TextEdit[] | null>(
		'textDocument/formatting',
		{
			textDocument: { uri: document.uri.toString() },
			options: { tabSize, insertSpaces },
		}
	);
	return edits ?? undefined;
}

async function applyFormatEdits(
	document: TextDocument,
	edits: TextEdit[] | undefined
): Promise<boolean> {
	if (!edits || edits.length === 0) {
		return false;
	}
	const we = new WorkspaceEdit();
	we.set(document.uri, edits);
	return workspace.applyEdit(we);
}

export function activate(context: ExtensionContext) {
	const config = workspace.getConfiguration('nora');
	const serverExe = getNoraExecutable(context);

	const serverOptions: ServerOptions = {
		command: serverExe,
		args: ['lsp'],
		options: { shell: false },
	};

	const outputChannel = window.createOutputChannel('Nora Language Server');
	outputChannel.show(true);

	const clientOptions: LanguageClientOptions = {
		outputChannel: outputChannel,
		traceOutputChannel: outputChannel,
		documentSelector: [{ scheme: 'file', language: 'nr' }],
		synchronize: {
			fileEvents: workspace.createFileSystemWatcher('**/*.nr'),
		},
	};

	client = new LanguageClient(
		'noraLanguageServer',
		'Nora Language Server',
		serverOptions,
		clientOptions
	);

	clientStarted = client.start();

	context.subscriptions.push(
		commands.registerCommand('nora.formatDocument', async () => {
			const editor = window.activeTextEditor;
			if (!editor || editor.document.languageId !== 'nr') {
				void window.showWarningMessage('Open a .nr file to format.');
				return;
			}
			const edits = await formatDocumentViaLsp(editor.document);
			if (!(await applyFormatEdits(editor.document, edits))) {
				void window.showWarningMessage(
					'Format failed. Fix parse errors in the file and try again.'
				);
			}
		})
	);

	context.subscriptions.push(
		commands.registerCommand('nora.formatWorkspace', async () => {
			const folder = workspace.workspaceFolders?.[0];
			if (!folder) {
				void window.showWarningMessage(
					'Open a folder workspace to format all .nr files.'
				);
				return;
			}
			try {
				const { stdout, stderr } = await execFileAsync(
					serverExe,
					['fmt', '-w'],
					{ cwd: folder.uri.fsPath, maxBuffer: 10 * 1024 * 1024 }
				);
				const msg = (stdout || stderr || '').trim();
				if (msg) {
					void window.showInformationMessage(msg);
				} else {
					void window.showInformationMessage('Workspace format complete.');
				}
			} catch (err: unknown) {
				const message =
					err instanceof Error ? err.message : String(err);
				void window.showErrorMessage(`nora fmt failed: ${message}`);
			}
		})
	);

	// Premium Concurrency Debugger (DAP Tracker)
	context.subscriptions.push(
		debug.registerDebugAdapterTrackerFactory('*', {
			createDebugAdapterTracker(session: DebugSession) {
				return {
					onWillReceiveMessage: m => {
						// Messages from VS Code TO the debug adapter
						if (m.type === 'request' && m.command === 'stepIn') {
							const editor = window.activeTextEditor;
							if (editor && editor.document.languageId === 'nr') {
								const line = editor.document.lineAt(editor.selection.active.line).text;
								if (line.includes('spawn') || line.includes('parallel')) {
									// Intercept stepIn and turn it into next (Step Over)
									m.command = 'next';
									// Set the global stepping flag in the C runtime
									session.customRequest('evaluate', { 
										expression: '__nora_stepping_into_spawn = 1', 
										context: 'hover' 
									}).then(() => {}, () => {});
								}
							}
						}

						if (m.type === 'request' && m.command === 'continue') {
							// Release scheduler locking on continue
							session.customRequest('evaluate', { 
								expression: '__nora_debug_locked_fiber = 0', 
								context: 'hover' 
							}).then(() => {}, () => {});
						}
					},
					onDidSendMessage: m => {
						// Messages from Debug Adapter TO VS Code
						if (m.type === 'event' && m.event === 'stopped') {
							// Automatically lock the scheduler to the currently active fiber when stopped at any breakpoint/step
							session.customRequest('evaluate', { 
								expression: '__nora_debug_locked_fiber = nr_fiber_current()', 
								context: 'hover' 
							}).then(() => {}, () => {});
						}
					}
				};
			}
		})
	);

	if (config.get<boolean>('format.onSave', false)) {
		context.subscriptions.push(
			workspace.onWillSaveTextDocument(async (event) => {
				if (event.document.languageId !== 'nr') {
					return;
				}
				const edits = await formatDocumentViaLsp(event.document);
				if (edits && edits.length > 0) {
					const we = new WorkspaceEdit();
					we.set(event.document.uri, edits);
					event.waitUntil(workspace.applyEdit(we));
				}
			})
		);
	}

	// Register debug hover provider to show both value and pointer address for heap allocated variables
	context.subscriptions.push(
		languages.registerHoverProvider(
			{ scheme: 'file', language: 'nr' },
			new NoraDebugHoverProvider()
		)
	);
}

export class NoraDebugHoverProvider implements HoverProvider {
	async provideHover(
		document: TextDocument,
		position: Position,
		token: CancellationToken
	): Promise<Hover | undefined> {
		const session = debug.activeDebugSession;
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
		let frameId: number | undefined = undefined;
		const debugAny = debug as any;
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
						const md = new MarkdownString();
						md.isTrusted = true;
						md.appendMarkdown(`### Nora Debugger Hover\n\n`);
						md.appendMarkdown(`* **Value**: \`${derefValue}\`\n`);
						md.appendMarkdown(`* **Address**: \`${addressOrValue}\`\n`);
						return new Hover(md, range);
					}
				}
			} catch {
				// Dereferencing failed - word is not a pointer. Fall back to default hover.
				return undefined;
			}
		} catch {
			// Evaluation failed
			return undefined;
		}

		return undefined;
	}
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}
