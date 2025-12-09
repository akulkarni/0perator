import {
  type ClientConfig,
  getSupportedClients as getMcpSupportedClients,
  supportedClients as mcpSupportedClients,
} from "./mcpInstall.js";

// Re-export types
export type { ClientConfig };

// ClientInfo for UI display (simplified interface)
export interface ClientInfo {
  name: string; // Client identifier (e.g., "claude-code")
  displayName: string; // Human-readable name (e.g., "Claude Code")
}

// Convert mcpInstall's supportedClients to the UI-friendly format
export const supportedClients: ClientInfo[] = mcpSupportedClients.map((c) => ({
  name: c.editorNames[0],
  displayName: c.name,
}));

export function getClientByName(name: string): ClientInfo | undefined {
  return supportedClients.find((c) => c.name === name);
}

// Re-export the full client configs for advanced use cases
export { getMcpSupportedClients, mcpSupportedClients };
