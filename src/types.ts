export interface ServerContext extends Record<string, unknown> {
  // No database connection needed for 0perator
  // Context can be extended later if needed
}
