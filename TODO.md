# 0perator TODO

## v0 - MVP (In Progress)

### ✅ Completed

- [x] Project scaffolding and structure
- [x] `go.mod` with dependencies
- [x] `0perator init` command
  - [x] Auto-install tiger-cli
  - [x] Tiger Cloud authentication
  - [x] Multi-IDE support (Claude Code, Cursor, Windsurf)
  - [x] Styled UI with #ff4f00 accent color
  - [x] Progress bars and timing
- [x] MCP server implementation
  - [x] Basic server using `mcp-go`
  - [x] `create_app` tool
  - [x] Tool definitions and handlers
- [x] Template system
  - [x] Generic scaffolding engine with embed.FS
  - [x] Template variable substitution
  - [x] Use-case first naming convention
- [x] `web-node` template
  - [x] package.json with modern dependencies
  - [x] TypeScript configuration (strict mode)
  - [x] Fastify server with Zod validation
  - [x] Conditional database integration
  - [x] Testing setup (Vitest)
  - [x] Linting (ESLint) and formatting (Prettier)
  - [x] Complete README

- [x] `deploy_local` MCP tool implementation
  - [x] Process management (start/stop/status)
  - [x] Port allocation (auto-assign or manual)
  - [x] Log streaming to files
  - [x] Health checks with retry logic
  - [x] Support for Node.js apps
  - [x] Additional tools: `stop_local`, `list_local`, `logs_local`
  - [x] Runtime package (`internal/runtime/process.go`)
  - [x] Dependency auto-installation (npm install)

### 📋 TODO for v0

- [x] End-to-end testing
  - [x] Test full flow: init → create_app → deploy_local
  - [x] Test in Claude Code
  - [x] MCP server successfully loads and tools are available
- [ ] Refine MCP tools
  - [ ] Test and improve create_app tool
  - [ ] Test and improve deploy_local tool
  - [ ] Test and improve stop_local tool
  - [ ] Test and improve list_local tool
  - [ ] Test and improve logs_local tool
  - [ ] Add better error messages
  - [ ] Validate input parameters
  - [ ] Handle edge cases
- [x] Improve installation DX
  - [x] Fix progress bar rendering issues (replaced with spinner)
  - [x] Better terminal output formatting (consistent 2-space indentation)
  - [x] Cleaner status messages (simplified text, better alignment)
  - [x] Fix IDE selection display order consistency (fixed order)
  - [x] Add colored output for success/error states (orange accent color 196)
  - [x] Cool ASCII art banner with brand colors
  - [x] Updated tagline: "Infrastructure for AI native development"
- [ ] Additional templates
  - [ ] `api-node` - REST API only (no frontend)
  - [ ] `cli-node` - CLI tool template
- [ ] Documentation
  - [ ] Installation guide
  - [ ] Usage examples
  - [ ] Template customization guide
- [x] Build and distribution
  - [x] Build script for multi-platform binaries
  - [x] Release process (GitHub Actions)
  - [x] Install script (scripts/install.sh)
  - [ ] Host install script at https://cli.0p.dev

## v1 - Expansion

### v1a - External Deploy
- [ ] Vercel deployment integration
- [ ] Environment variable management
- [ ] Production URL handling

### v1b - CLI Commands
- [ ] `0perator create app` - Direct CLI usage
- [ ] `0perator deploy` - Deploy without IDE
- [ ] `0perator status` - Show running apps
- [ ] `0perator logs` - View app logs

### v1c - Multi-agent Support
- [ ] Parallel app creation
- [ ] Workspace management
- [ ] Inter-app communication

### v1d - Pricing
- [ ] Usage tracking
- [ ] Paid tier features
- [ ] Billing integration

## v2 - Sharing / Discovery
- [ ] App gallery
- [ ] Public/private apps
- [ ] Share app URLs
- [ ] Fork/remix functionality

## v3 - Prompt Templates
- [ ] Create prompt template system (like Tiger MCP)
- [ ] Common app patterns as prompts
- [ ] Community-contributed prompts

## v4 - Day 2 Operations
- [ ] "Why is my app slow?" diagnostics
- [ ] Database query analysis
- [ ] Performance monitoring
- [ ] Error tracking

## v5 - Stripe Integration
- [ ] Automatic billing setup
- [ ] Payment forms
- [ ] Webhook handling
- [ ] Subscription management

## Technical Debt / Improvements

- [ ] Error handling improvements
- [ ] Better logging
- [ ] Configuration file support (~/.config/0perator/config.json)
- [ ] Update check on startup
- [ ] Telemetry/analytics (opt-in)
- [ ] Refactor tiger-cli dependency
  - [ ] Option: Import as library (requires tiger-cli changes)
  - [ ] Option: Keep shelling out (current approach)

## Questions / Decisions

- [ ] How to handle template updates?
- [ ] Should we support custom user templates?
- [ ] Database migration strategy for templates?
- [ ] How to handle breaking changes in templates?
