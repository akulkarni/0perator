import { exec } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { homedir } from "node:os";
import { dirname, join } from "node:path";
import { promisify } from "node:util";
import { parse, stringify } from "comment-json";

const execAsync = promisify(exec);

// MCPServerConfig represents the MCP server configuration
export interface MCPServerConfig {
  command: string;
  args: string[];
}

// InstallOptions configures the MCP server installation behavior
export interface InstallOptions {
  // ClientName is the name of the client to configure (required)
  clientName: string;
  // ServerName is the name to register the MCP server as (required)
  serverName: string;
  // Command is the path to the MCP server binary (required)
  command: string;
  // Args are the arguments to pass to the MCP server binary (required)
  args: string[];
  // CreateBackup creates a backup of existing config files before modification
  createBackup?: boolean;
  // CustomConfigPath overrides the default config file location
  customConfigPath?: string;
}

// ClientConfig represents our own client configuration for MCP installation
export interface ClientConfig {
  name: string;
  editorNames: string[];
  mcpServersPathPrefix?: string; // JSON path prefix for MCP servers config
  configPaths: string[];
  buildInstallCommand?: (
    serverName: string,
    command: string,
    args: string[],
  ) => string[] | null;
}

// ClientInfo contains information about a supported MCP client
export interface ClientInfo {
  name: string;
  clientName: string;
}

// Supported clients configuration
export const supportedClients: ClientConfig[] = [
  {
    name: "Claude Code",
    editorNames: ["claude-code"],
    configPaths: ["~/.claude.json"],
    buildInstallCommand: (serverName, command, args) => [
      "claude",
      "mcp",
      "add",
      "-s",
      "user",
      serverName,
      command,
      ...args,
    ],
  },
  {
    name: "Cursor",
    editorNames: ["cursor"],
    mcpServersPathPrefix: "/mcpServers",
    configPaths: ["~/.cursor/mcp.json"],
  },
  {
    name: "Windsurf",
    editorNames: ["windsurf"],
    mcpServersPathPrefix: "/mcpServers",
    configPaths: ["~/.codeium/windsurf/mcp_config.json"],
  },
  {
    name: "Codex",
    editorNames: ["codex"],
    configPaths: ["~/.codex/config.toml", "$CODEX_HOME/config.toml"],
    buildInstallCommand: (serverName, command, args) => [
      "codex",
      "mcp",
      "add",
      serverName,
      command,
      ...args,
    ],
  },
  {
    name: "Gemini CLI",
    editorNames: ["gemini", "gemini-cli"],
    configPaths: ["~/.gemini/settings.json"],
    buildInstallCommand: (serverName, command, args) => [
      "gemini",
      "mcp",
      "add",
      "-s",
      "user",
      serverName,
      command,
      ...args,
    ],
  },
  {
    name: "VS Code",
    editorNames: ["vscode", "code", "vs-code"],
    configPaths: [
      "~/.config/Code/User/mcp.json",
      "~/Library/Application Support/Code/User/mcp.json",
      "~/AppData/Roaming/Code/User/mcp.json",
    ],
    buildInstallCommand: (serverName, command, args) => {
      const config = JSON.stringify({
        name: serverName,
        command: command,
        args: args,
      });
      return ["code", "--add-mcp", config];
    },
  },
  {
    name: "Google Antigravity",
    editorNames: ["antigravity", "agy"],
    mcpServersPathPrefix: "/mcpServers",
    configPaths: ["~/.gemini/antigravity/mcp_config.json"],
  },
  {
    name: "Kiro CLI",
    editorNames: ["kiro-cli"],
    configPaths: ["~/.kiro/settings/mcp.json"],
    buildInstallCommand: (serverName, command, args) => [
      "kiro-cli",
      "mcp",
      "add",
      "--name",
      serverName,
      "--command",
      command,
      "--args",
      args.join(","),
    ],
  },
];

/**
 * Expand path by replacing ~ with home directory and environment variables
 */
export function expandPath(path: string): string {
  let expanded = path;

  // Expand ~ to home directory
  if (expanded.startsWith("~/")) {
    expanded = join(homedir(), expanded.slice(2));
  } else if (expanded === "~") {
    expanded = homedir();
  }

  // Expand environment variables like $VAR or ${VAR}
  expanded = expanded.replace(/\$\{?(\w+)\}?/g, (_, varName) => {
    return process.env[varName] || "";
  });

  return expanded;
}

/**
 * Get all valid editor names from supported clients
 */
export function getValidEditorNames(): string[] {
  const names: string[] = [];
  for (const client of supportedClients) {
    names.push(...client.editorNames);
  }
  return names;
}

/**
 * Get information about all supported MCP clients
 */
export function getSupportedClients(): ClientInfo[] {
  return supportedClients.map((c) => ({
    name: c.name,
    clientName: c.editorNames[0],
  }));
}

/**
 * Find client configuration by name
 */
export function findClientConfig(clientName: string): ClientConfig | null {
  const normalizedName = clientName.toLowerCase();

  for (const client of supportedClients) {
    for (const name of client.editorNames) {
      if (name.toLowerCase() === normalizedName) {
        return client;
      }
    }
  }

  return null;
}

/**
 * Find client configuration file from a list of possible paths
 */
export function findClientConfigFile(configPaths: string[]): string | null {
  for (const path of configPaths) {
    const expandedPath = expandPath(path);
    if (existsSync(expandedPath)) {
      return expandedPath;
    }
  }

  // If no existing config found, use the first path as default
  if (configPaths.length > 0) {
    return expandPath(configPaths[0]);
  }

  return null;
}

/**
 * Create a backup of the configuration file
 */
export function createConfigBackup(configPath: string): string | null {
  if (!existsSync(configPath)) {
    return null;
  }

  const backupPath = `${configPath}.backup.${Date.now()}`;
  const content = readFileSync(configPath);
  const mode = statSync(configPath).mode;

  writeFileSync(backupPath, content, { mode });

  return backupPath;
}

/**
 * Add MCP server using CLI command
 */
async function addMCPServerViaCLI(
  clientCfg: ClientConfig,
  serverName: string,
  command: string,
  args: string[],
): Promise<void> {
  if (!clientCfg.buildInstallCommand) {
    throw new Error(
      `No install command configured for client ${clientCfg.name}`,
    );
  }

  const installCommand = clientCfg.buildInstallCommand(
    serverName,
    command,
    args,
  );
  if (!installCommand) {
    throw new Error(`Failed to build install command for ${clientCfg.name}`);
  }

  const [cmd, ...cmdArgs] = installCommand;

  try {
    await execAsync(`${cmd} ${cmdArgs.map((a) => `"${a}"`).join(" ")}`);
  } catch (err) {
    const error = err as Error & { stderr?: string; stdout?: string };
    const output = error.stderr || error.stdout || "";
    throw new Error(
      `Failed to run ${clientCfg.name} installation command: ${error.message}${output ? `\nOutput: ${output}` : ""}`,
    );
  }
}

/**
 * Add MCP server via JSON configuration file
 * Uses comment-json to preserve comments in the config file
 */
export function addMCPServerViaJSON(
  configPath: string,
  mcpServersPathPrefix: string,
  serverName: string,
  command: string,
  args: string[],
): void {
  // Create configuration directory if it doesn't exist
  const configDir = dirname(configPath);
  if (!existsSync(configDir)) {
    mkdirSync(configDir, { recursive: true });
  }

  // MCP server configuration
  const serverConfig: MCPServerConfig = {
    command,
    args,
  };

  // Get original file mode or use default
  let fileMode = 0o600;
  if (existsSync(configPath)) {
    fileMode = statSync(configPath).mode;
  }

  // Read existing configuration or create empty one
  // Using comment-json to preserve comments
  let config: Record<string, unknown> = {};
  if (existsSync(configPath)) {
    const content = readFileSync(configPath, "utf-8");
    if (content.trim()) {
      try {
        config = parse(content) as Record<string, unknown>;
      } catch {
        throw new Error(`Failed to parse existing config at ${configPath}`);
      }
    }
  }

  // Navigate to the mcpServers path and create if needed
  // mcpServersPathPrefix is like "/mcpServers"
  const pathParts = mcpServersPathPrefix.split("/").filter((p) => p);

  let current: Record<string, unknown> = config;
  for (const part of pathParts) {
    if (!(part in current)) {
      current[part] = {};
    }
    current = current[part] as Record<string, unknown>;
  }

  // Add the server configuration
  current[serverName] = serverConfig;

  // Write back to file, preserving comments
  writeFileSync(configPath, `${stringify(config, null, 2)}\n`, {
    mode: fileMode,
  });
}

/**
 * Install MCP server configuration for the specified client
 * This is the main installation function that handles both CLI and JSON-based installation
 */
export async function installMCPForClient(opts: InstallOptions): Promise<void> {
  // Validate required options
  if (!opts.clientName) {
    throw new Error("clientName is required");
  }
  if (!opts.serverName) {
    throw new Error("serverName is required");
  }
  if (!opts.command) {
    throw new Error("command is required");
  }
  if (!opts.args) {
    throw new Error("args is required");
  }

  // Find the client configuration by name
  const clientCfg = findClientConfig(opts.clientName);
  if (!clientCfg) {
    const supportedNames = getValidEditorNames();
    throw new Error(
      `Unsupported client: ${opts.clientName}. Supported clients: ${supportedNames.join(", ")}`,
    );
  }

  const mcpServersPathPrefix = clientCfg.mcpServersPathPrefix;

  let configPath: string | null = null;
  if (opts.customConfigPath) {
    configPath = expandPath(opts.customConfigPath);
  } else if (clientCfg.configPaths.length > 0) {
    configPath = findClientConfigFile(clientCfg.configPaths);
  } else if (!clientCfg.buildInstallCommand) {
    throw new Error(
      `Client ${opts.clientName} has no configPaths or buildInstallCommand defined`,
    );
  }

  // Create backup if requested and we have a config file
  if (opts.createBackup && configPath && existsSync(configPath)) {
    createConfigBackup(configPath);
  }

  // Add MCP server to configuration
  if (clientCfg.buildInstallCommand) {
    // Use CLI approach when install command builder is configured
    await addMCPServerViaCLI(
      clientCfg,
      opts.serverName,
      opts.command,
      opts.args,
    );
  } else {
    // Use JSON patching approach for JSON-config clients
    if (!configPath) {
      throw new Error(`No config path found for ${opts.clientName}`);
    }
    if (!mcpServersPathPrefix) {
      throw new Error(
        `No MCP servers path prefix configured for ${opts.clientName}`,
      );
    }
    addMCPServerViaJSON(
      configPath,
      mcpServersPathPrefix,
      opts.serverName,
      opts.command,
      opts.args,
    );
  }
}
