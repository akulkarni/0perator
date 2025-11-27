# TODO

## Refactoring

- [ ] Create a shared library with tiger-cli for MCP install/uninstall logic
  - Both 0perator and tiger-cli need to manage MCP server configurations
  - IDE config paths, config file parsing, and server registration should be in a common package
  - This would ensure consistent behavior and reduce duplication across projects
