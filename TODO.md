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
  - [x] 3-tool architecture (discover_patterns, get_template, execute)
  - [x] Tool definitions and handlers
- [x] Prompt template system
  - [x] Frontmatter parser with YAML
  - [x] Tag-based discovery with scoring
  - [x] Category defaults system
  - [x] Template loading from embed.FS
  - [x] Semantic search algorithm
- [x] Execute primitives
  - [x] run_command - Shell command execution
  - [x] read_file - Read file contents
  - [x] create_file - Create/overwrite files
  - [x] edit_file - Patch existing files
  - [x] start_process - Start and track processes
  - [x] stop_process - Stop tracked processes
  - [x] get_logs - Get process logs
  - [x] list_processes - List running processes
- [x] Process management
  - [x] Local deployment (bare processes)
  - [x] Port allocation (auto-assign or manual)
  - [x] Log streaming to files
  - [x] Health checks with retry logic
  - [x] Runtime package (`internal/runtime/process.go`)
  - [x] Dependency auto-installation (npm install)
- [x] Initial templates
  - [x] `create_web_app.md` - Comprehensive Node.js + TypeScript + Fastify guide
- [x] Improve installation DX
  - [x] Fix progress bar rendering issues (replaced with spinner)
  - [x] Better terminal output formatting (consistent 2-space indentation)
  - [x] Cleaner status messages (simplified text, better alignment)
  - [x] Fix IDE selection display order consistency (fixed order)
  - [x] Add colored output for success/error states (orange accent color 196)
  - [x] Cool ASCII art banner with brand colors
  - [x] Updated tagline: "Infrastructure for AI native development"
- [x] Build and distribution
  - [x] Build script for multi-platform binaries
  - [x] Release process (GitHub Actions)
  - [x] Install script (scripts/install.sh)
- [x] Documentation
  - [x] Updated README with new architecture
  - [x] Architecture diagrams
  - [x] Workflow examples

### ✅ Completed for v0

- [x] Complete all 7 foundational templates
  - [x] `create_web_app.md` - Node.js + TypeScript + Fastify web applications
  - [x] `database_tiger.md` - Tiger Cloud PostgreSQL/TimescaleDB integration
  - [x] `auth_jwt.md` - JWT authentication with email verification and password reset
  - [x] `email_resend.md` - Resend email integration for transactional emails
  - [x] `payments_stripe.md` - Stripe payment integration (one-time + subscriptions)
  - [x] `deploy_cloudflare.md` - Cloudflare Pages for static sites and serverless functions
  - [x] `deploy_railway.md` - Railway for Node.js servers and long-running workloads
- [x] Configure defaults
  - [x] Set default templates in `internal/prompts/defaults.go`
  - [x] All 5 categories configured (deployment, database, authentication, payments, email)
- [x] Testing
  - [x] Template loading tests (all 7 templates load correctly)
  - [x] Default configuration tests (all 5 defaults verified)
  - [x] Discovery algorithm tests (tag-based search working)

### 📋 TODO for v0.2.0 Release

- [ ] Real-world testing
  - [ ] Test in Claude Code with actual app creation
  - [ ] Test template composition (web → db → auth → payments → deploy)
  - [ ] Verify all 8 execute primitives work in practice
  - [ ] Test in Cursor/Windsurf (if possible)
- [ ] Refinements
  - [ ] Better error messages in execute operations
  - [ ] Input validation for all operations
  - [ ] Handle edge cases (port conflicts, missing files, etc.)
- [ ] Distribution
  - [ ] Host install script at https://cli.0p.dev
  - [ ] Test installation on clean machines
  - [ ] Create release (v0.2.0)
  - [ ] Update README with examples

## v1 - Expansion

### v1a - Additional Templates
- [ ] `create_api.md` - REST API only (no frontend)
- [ ] `create_cli.md` - CLI tool template
- [ ] `email_sendgrid.md` - SendGrid email integration
- [ ] `storage_r2.md` - Cloudflare R2 file storage
- [ ] `storage_s3.md` - AWS S3 file storage
- [ ] `testing_vitest.md` - Vitest testing setup
- [ ] `feature_websockets.md` - Real-time WebSocket features
- [ ] `feature_cron.md` - Scheduled tasks

### v1b - External Deploy
- [ ] `deploy_vercel.md` - Vercel deployment
- [ ] `deploy_railway.md` - Railway deployment
- [ ] `deploy_fly.md` - Fly.io deployment
- [ ] Environment variable management
- [ ] Production URL handling

### v1c - Template Enhancements
- [ ] Template versioning
- [ ] User-configurable defaults (~/.config/0perator/config.yaml)
- [ ] Template validation tool
- [ ] Community template contributions
- [ ] Template marketplace infrastructure

### v1d - CLI Commands
- [ ] `0perator templates list` - List available templates
- [ ] `0perator templates search <query>` - Search templates
- [ ] `0perator status` - Show running apps
- [ ] `0perator logs <process-id>` - View app logs

## v2 - Discovery & Sharing
- [ ] Public template registry
- [ ] Template versioning and updates
- [ ] Community-contributed templates
- [ ] Template ratings and reviews
- [ ] Fork/remix functionality

## v3 - Advanced Features
- [ ] Multi-agent support (parallel app creation)
- [ ] Workspace management
- [ ] Inter-app communication
- [ ] Template composition patterns

## v4 - Day 2 Operations
- [ ] "Why is my app slow?" diagnostics
- [ ] Database query analysis
- [ ] Performance monitoring
- [ ] Error tracking integration
- [ ] Log aggregation

## v5 - Enterprise
- [ ] Usage tracking
- [ ] Paid tier features
- [ ] Billing integration
- [ ] Team workspaces
- [ ] Access control

## Technical Debt / Improvements

- [ ] Error handling improvements across all operations
- [ ] Better logging (structured logging)
- [ ] Configuration file validation
- [ ] Update check on startup
- [ ] Telemetry/analytics (opt-in)
- [ ] Refactor tiger-cli dependency
  - [ ] Option: Import as library (requires tiger-cli changes)
  - [ ] Option: Keep shelling out (current approach)
- [ ] Performance optimization
  - [ ] Cache template loading
  - [ ] Parallel file operations
  - [ ] Faster discovery algorithm

## Architecture Decisions

### ✅ Resolved
- [x] **Architecture:** Prompt templates vs scaffolding → Prompt templates
- [x] **Tools:** Many specific tools vs few generic tools → 3 tools (discover/get/execute)
- [x] **Primitives:** 8 core operations for maximum flexibility
- [x] **Discovery:** Tag-based semantic search with scoring
- [x] **Defaults:** Category-based defaults (e.g., Cloudflare for deployment)

### 🤔 Open Questions
- [ ] How to handle template updates when users have customized apps?
- [ ] Should we support custom user templates? How?
- [ ] Database migration strategy for template-generated apps?
- [ ] How to handle breaking changes in templates?
- [ ] Should templates include "upgrade paths" to newer versions?
- [ ] How to handle multiple versions of the same template?
