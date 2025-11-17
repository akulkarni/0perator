# 0perator Action-Based Architecture Implementation Plan

## Executive Summary

Transition from template-based system to UNIX-inspired action-based architecture with recipes, maintaining backward compatibility while achieving O(N) complexity and sub-minute execution times for common patterns.

**Key Principles:**
- Small actions that do one thing well (UNIX philosophy)
- Recipes compose actions (like shell pipelines)
- Single operator tool interface (not 50+ tools)
- LLM handles orchestration naturally

## Phase 1: Foundation (Week 1)
### Goal: Build core infrastructure without breaking existing system

### 1.1 Action Registry System
```go
// internal/actions/registry.go
type Action struct {
    Name         string
    Description  string
    Category     string   // "create", "setup", "add", "deploy"
    Inputs       []Input
    Outputs      []Output
    Dependencies []string  // Other actions that must run first
    Conflicts    []string  // Actions that cannot coexist
    Tier         string   // "fast" or "flexible"
    Implementation func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

type Registry struct {
    actions map[string]*Action
    graph   *DependencyGraph
}
```

### 1.2 Action Definition Format
Create directory structure:
```
internal/
  actions/
    definitions/
      create/
        web_app.yaml
        api_server.yaml
      setup/
        postgres.yaml
        redis.yaml
      add/
        jwt_auth.yaml
        stripe_payments.yaml
    implementations/
      create_web_app.go
      setup_postgres.go
```

Example action definition:
```yaml
# internal/actions/definitions/create/web_app.yaml
name: create_web_app
description: "Create a new web application with specified framework"
category: create
tier: fast
estimated_time: 15s
inputs:
  - name: framework
    type: string
    required: true
    options: ["nextjs", "react", "vue", "svelte"]
    default: "nextjs"
  - name: directory
    type: string
    required: false
    default: "."
  - name: typescript
    type: boolean
    default: true
outputs:
  - name: project_path
    type: string
  - name: package_json_path
    type: string
dependencies: []
conflicts: []
```

### 1.3 Operator Tool Interface
```go
// internal/operator/operator.go
type Operator struct {
    registry *actions.Registry
    executor *Executor
    cache    *ActionCache
}

func (o *Operator) ExecuteAction(name string, inputs map[string]interface{}) (map[string]interface{}, error)
func (o *Operator) ExecuteSequence(actions []ActionCall) (*ExecutionResult, error)
func (o *Operator) ValidateSequence(actions []ActionCall) error
func (o *Operator) GetAvailableActions(category string) []ActionMetadata
```

### Deliverables:
- [ ] Action registry implementation
- [ ] Action definition schema
- [ ] Operator tool scaffold
- [ ] Dependency graph resolver

## Phase 2: Core Actions (Week 2)
### Goal: Convert most common templates to actions

### 2.1 Priority Actions to Implement
Based on usage patterns, implement these first:

**Create Actions** (Foundation):
- `create_web_app` - Next.js/React/Vue apps
- `create_api_server` - Express/Fastify/Hono servers
- `create_cli_tool` - CLI applications
- `create_directory_structure` - Project scaffolding

**Setup Actions** (Infrastructure):
- `setup_postgres` - PostgreSQL with TimescaleDB
- `setup_redis` - Redis cache
- `setup_docker_compose` - Container orchestration
- `setup_env_file` - Environment configuration

**Add Actions** (Features):
- `add_jwt_auth` - JWT authentication
- `add_oauth` - OAuth providers
- `add_stripe_payments` - Payment processing
- `add_user_model` - User database schema
- `add_api_endpoint` - REST endpoints
- `add_middleware` - Express/Next.js middleware

**Deploy Actions** (Deployment):
- `deploy_local` - Local development server
- `deploy_railway` - Railway deployment
- `deploy_vercel` - Vercel deployment
- `deploy_cloudflare` - Cloudflare Workers

### 2.2 Action Implementation Pattern
```go
// internal/actions/implementations/create_web_app.go
func CreateWebApp(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    framework := inputs["framework"].(string)
    directory := inputs["directory"].(string)

    // Fast, deterministic implementation
    switch framework {
    case "nextjs":
        return createNextApp(ctx, directory, inputs)
    case "react":
        return createReactApp(ctx, directory, inputs)
    // ...
    }
}

func createNextApp(ctx context.Context, dir string, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Direct implementation, no template interpretation
    cmd := exec.Command("npx", "create-next-app@latest", dir, "--typescript", "--tailwind", "--app")
    // ...

    return map[string]interface{}{
        "project_path": filepath.Abs(dir),
        "package_json_path": filepath.Join(dir, "package.json"),
    }, nil
}
```

### Deliverables:
- [ ] 20 core actions implemented
- [ ] Unit tests for each action
- [ ] Performance benchmarks
- [ ] Action documentation

## Phase 3: Recipe System (Week 2-3)
### Goal: Enable powerful composition through recipes

### 3.1 Recipe Format
```yaml
# internal/recipes/saas_starter.yaml
name: saas_starter
description: "Production-ready SaaS application with auth and payments"
tags: ["saas", "fullstack", "production"]
category: complete_app
inputs:
  - name: app_name
    type: string
    required: true
  - name: database
    type: string
    default: "postgres"
    options: ["postgres", "mysql", "sqlite"]
actions:
  - action: create_web_app
    inputs:
      framework: "nextjs"
      directory: "{{app_name}}"
      typescript: true
  - action: setup_postgres
    inputs:
      database_name: "{{app_name}}_db"
      with_timescale: true
  - action: add_jwt_auth
    inputs:
      secret_source: "env"
  - action: add_user_model
    inputs:
      fields: ["email", "name", "created_at"]
  - action: add_stripe_payments
    inputs:
      products: ["monthly", "yearly"]
  - action: deploy_local
    inputs:
      port: 3000
```

### 3.2 Recipe Engine
```go
// internal/recipes/engine.go
type RecipeEngine struct {
    recipes  map[string]*Recipe
    operator *operator.Operator
}

func (r *RecipeEngine) ExecuteRecipe(name string, inputs map[string]interface{}) (*RecipeResult, error) {
    recipe := r.recipes[name]

    // Validate inputs
    if err := recipe.ValidateInputs(inputs); err != nil {
        return nil, err
    }

    // Resolve action sequence with inputs
    sequence := recipe.ResolveActions(inputs)

    // Execute through operator
    return r.operator.ExecuteSequence(sequence)
}
```

### 3.3 Common Recipes
- `saas_starter` - Full SaaS with auth, payments, database
- `api_backend` - REST API with database and auth
- `static_site` - Static website with deployment
- `enterprise_app` - Enterprise features (SSO, audit logs, RBAC)
- `ai_app` - AI application with vector DB and LLM integration
- `ecommerce` - E-commerce with cart, checkout, inventory

### Deliverables:
- [ ] Recipe schema definition
- [ ] Recipe engine implementation
- [ ] 10+ common recipes
- [ ] Recipe validation system

## Phase 4: MCP Integration (Week 3)
### Goal: Replace three tools with single operator interface

### 4.1 New MCP Tool Structure
```go
// internal/server/tools_v2.go
func (s *Server) registerToolsV2() {
    // Single operator tool replaces all three
    mcp.AddTool(s.mcpServer, &mcp.Tool{
        Name: "operator",
        Description: "Execute actions and recipes to build applications. Use 'discover' to find available actions/recipes, 'execute' to run them.",
    }, s.handleOperator)
}

type OperatorInput struct {
    Command string                 `json:"command"` // "discover", "execute", "validate"
    Type    string                 `json:"type"`    // "action", "recipe", "sequence"
    Name    string                 `json:"name,omitempty"`
    Query   string                 `json:"query,omitempty"`
    Inputs  map[string]interface{} `json:"inputs,omitempty"`
    Actions []ActionCall           `json:"actions,omitempty"`
}
```

### 4.2 LLM Integration Pattern
The LLM will:
1. Parse user intent
2. Discover relevant actions/recipes
3. Compose execution plan
4. Execute through operator
5. Report results

Example flow:
```
User: "Build a SaaS app with auth and Stripe"
LLM: operator.discover(type="recipe", query="saas stripe auth")
Response: [saas_starter, enterprise_saas, ...]
LLM: operator.execute(type="recipe", name="saas_starter", inputs={...})
Response: Success, app running at localhost:3000
```

### Deliverables:
- [ ] Single operator MCP tool
- [ ] Discovery interface
- [ ] Execution interface
- [ ] Progress streaming

## Phase 5: Migration & Testing (Week 4)
### Goal: Seamless transition with no breaking changes

### 5.1 Template-to-Action Converter
```go
// internal/migration/converter.go
func ConvertTemplateToActions(templatePath string) ([]Action, error) {
    // Parse existing template
    // Extract discrete operations
    // Map to action definitions
    // Generate action implementations
}
```

### 5.2 Backward Compatibility Layer
Keep existing tools working during transition:
```go
// Support both old and new simultaneously
func (s *Server) handleGetTemplate(...) {
    // If action-based version exists, convert to template format
    // Otherwise, return original template
}
```

### 5.3 Testing Strategy
- **Unit Tests**: Each action in isolation
- **Integration Tests**: Common recipes end-to-end
- **Performance Tests**: Target < 1 minute for saas_starter
- **Compatibility Tests**: Old templates still work
- **LLM Tests**: Natural language → correct actions

### Deliverables:
- [ ] Migration tooling
- [ ] Test suite (100+ tests)
- [ ] Performance benchmarks
- [ ] Rollback plan

## Phase 6: Optimization (Week 5)
### Goal: Achieve sub-minute execution for common patterns

### 6.1 Caching Strategy
```go
// internal/cache/cache.go
type ActionCache struct {
    // Cache package.json resolution
    // Cache npm install results
    // Cache docker images
    // Cache action outputs when deterministic
}
```

### 6.2 Parallel Execution
```go
// Execute independent actions concurrently
func (e *Executor) ExecuteParallel(actions []ActionCall) {
    // Build dependency graph
    // Find independent action sets
    // Execute in parallel waves
}
```

### 6.3 Fast Path Optimizations
- Pre-pull Docker images
- Cache npm packages locally
- Template file generation (no network)
- Batch database operations

### Deliverables:
- [ ] Caching system
- [ ] Parallel executor
- [ ] Performance monitoring
- [ ] < 1 minute for common recipes

## Phase 7: Community & Launch (Week 6)
### Goal: Enable community contributions

### 7.1 Contribution Guide
```markdown
# Contributing Actions to 0perator

## Creating an Action
1. Define action in YAML
2. Implement in Go
3. Add tests
4. Submit PR

## Creating a Recipe
1. Compose existing actions
2. Define inputs/outputs
3. Test end-to-end
4. Submit PR
```

### 7.2 Developer Tools
```bash
# CLI for action development
0perator dev action create my_action
0perator dev action test my_action
0perator dev recipe validate my_recipe
0perator dev benchmark saas_starter
```

### 7.3 Documentation
- Action catalog with examples
- Recipe cookbook
- Architecture guide
- Migration guide from templates

### Deliverables:
- [ ] Contribution guidelines
- [ ] Developer CLI tools
- [ ] Documentation site
- [ ] Example actions/recipes

## Success Metrics

### Performance
- [Target] Common recipes < 1 minute
- [Target] Simple actions < 10 seconds
- [Target] 95% success rate

### Scale
- [Target] 50+ actions without complexity issues
- [Target] 20+ recipes covering common use cases
- [Target] Support 100+ action compositions

### Developer Experience
- [Target] Natural language → working app
- [Target] Clear action execution visibility
- [Target] Helpful error messages

### Contributor Experience
- [Target] < 30 minutes to contribute first action
- [Target] < 1 hour to create complex recipe
- [Target] 10+ community contributions in first month

## Risk Mitigation

### Risk: Breaking existing workflows
**Mitigation**: Maintain backward compatibility, gradual rollout

### Risk: Performance regression
**Mitigation**: Benchmark every change, cache aggressively

### Risk: Too complex for contributors
**Mitigation**: Excellent docs, examples, dev tools

### Risk: LLM confusion with new pattern
**Mitigation**: Clear tool descriptions, progressive disclosure

## Rollout Plan

### Week 1-2: Build & Test Internally
- Core infrastructure
- Essential actions
- Internal testing

### Week 3-4: Alpha Testing
- Select users test new system
- Gather feedback
- Fix issues

### Week 5: Beta Release
- Public beta with opt-in flag
- Monitor performance
- Refine based on usage

### Week 6: General Availability
- Default to action-based system
- Maintain template fallback
- Launch community contributions

## Conclusion

This plan transforms 0perator from O(N²) template complexity to O(N) action composition, achieving:
- **Speed**: < 1 minute for production apps
- **Scale**: Unlimited actions without complexity explosion
- **Simplicity**: UNIX philosophy - small tools, powerful composition
- **Community**: Easy contributions via actions and recipes

The key insight: **Actions + Recipes + Single Operator = Magic**