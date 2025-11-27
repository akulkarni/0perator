# 0perator Development Guide

## Cobra CLI Architecture

This project uses Cobra with a **pure functional builder pattern** and **zero global command state**.

### Philosophy

- **No global variables** - All commands, flags, and state are locally scoped
- **Functional builders** - Every command is built by a dedicated `buildXXXCmd()` function
- **Complete tree building** - `buildRootCmd()` constructs the entire CLI structure
- **Perfect test isolation** - Each test gets completely fresh command instances

### Command Structure

```
buildRootCmd() → Complete CLI with all commands
├── buildVersionCmd()
├── buildInitCmd()
├── buildUninstallCmd()
└── buildMCPCmd()
```

### SilenceUsage Pattern

Control when usage information is displayed on errors:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    // 1. Do argument validation first - errors here show usage
    if len(args) < 1 {
        return fmt.Errorf("argument required")
    }

    // 2. Set SilenceUsage = true after argument validation
    cmd.SilenceUsage = true

    // 3. Proceed with business logic - errors here don't show usage
    if err := someOperation(); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }

    return nil
},
```

**Philosophy**:
- Early argument/syntax errors → show usage (helps users learn command syntax)
- Operational errors after validation → don't show usage (avoids cluttering output)

### Adding New Commands

1. Create a builder function in `internal/cmd/`:

```go
func buildMyCmd() *cobra.Command {
    // Declare flag variables locally (NEVER globally)
    var myFlag string

    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "Short description",
        Long:  `Longer description...`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Argument validation here (before SilenceUsage)

            cmd.SilenceUsage = true

            // Business logic here
            return nil
        },
    }

    // Add flags bound to local variables
    cmd.Flags().StringVar(&myFlag, "flag", "", "Flag description")

    return cmd
}
```

2. Add to `buildRootCmd()` in `internal/cmd/root.go`:

```go
cmd.AddCommand(buildMyCmd())
```

### Testing Commands

```go
func executeCommand(args ...string) (string, error) {
    rootCmd := buildRootCmd()

    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetErr(buf)
    rootCmd.SetArgs(args)

    err := rootCmd.Execute()
    return buf.String(), err
}

func TestMyCommand(t *testing.T) {
    output, err := executeCommand("mycommand", "--flag", "value")
    // assertions...
}
```

### Key Files

- `internal/cmd/root.go` - Root command builder and `Execute()` entry point
- `internal/cmd/*.go` - Individual command builders
- `cmd/0perator-mcp/main.go` - Main entry point (just calls `cmd.Execute()`)
