import { ClientInfo } from '../types.js';

export const supportedClients: ClientInfo[] = [
  { name: 'claude-code', displayName: 'Claude Code' },
  { name: 'cursor', displayName: 'Cursor' },
  { name: 'windsurf', displayName: 'Windsurf' },
];

export function getClientByName(name: string): ClientInfo | undefined {
  return supportedClients.find((c) => c.name === name);
}
