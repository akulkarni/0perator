export interface ServerContext extends Record<string, unknown> {
  // No database connection needed for 0perator
  // Context can be extended later if needed
}

export interface ClientInfo {
  name: string;
  displayName: string;
}
