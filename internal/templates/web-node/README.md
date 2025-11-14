# {{.AppName}}

{{.Description}}

## Quick Start

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## Development

```bash
# Run tests
npm test

# Run tests with UI
npm test:ui

# Type check
npm run typecheck

# Lint
npm run lint

# Format code
npm run format
```

## Environment Variables

```bash
PORT=3000                    # Server port (default: 3000)
HOST=0.0.0.0                # Server host (default: 0.0.0.0)
{{if .DatabaseURL}}
DATABASE_URL={{.DatabaseURL}}  # PostgreSQL connection string
{{end}}
```

## API Endpoints

- `GET /health` - Health check
- `GET /` - API info
- `POST /api/items` - Create item (example endpoint)

## Tech Stack

- **Runtime**: Node.js 20+
- **Framework**: Fastify
- **Language**: TypeScript
- **Validation**: Zod
{{if .DatabaseURL}}
- **Database**: PostgreSQL via postgres.js
{{end}}
- **Testing**: Vitest
- **Linting**: ESLint
- **Formatting**: Prettier

## Project Structure

```
{{.AppName}}/
├── src/
│   └── index.ts           # Main application entry point
├── tests/                 # Test files
├── dist/                  # Compiled output (generated)
├── package.json
├── tsconfig.json
└── README.md
```

---

Generated with [0perator](https://0p.dev)
