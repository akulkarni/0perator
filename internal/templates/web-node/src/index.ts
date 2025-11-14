import Fastify from 'fastify';
import cors from '@fastify/cors';
import { z } from 'zod';
{{if .DatabaseURL}}
import postgres from 'postgres';
{{end}}

const fastify = Fastify({
  logger: true,
});

await fastify.register(cors);

{{if .DatabaseURL}}
// Database connection
const sql = postgres('{{.DatabaseURL}}');

// Test database connection
try {
  await sql`SELECT 1`;
  fastify.log.info('Database connected successfully');
} catch (err) {
  fastify.log.error({ err }, 'Failed to connect to database');
  process.exit(1);
}
{{end}}

// Health check endpoint
fastify.get('/health', async () => {
  return { status: 'ok', timestamp: new Date().toISOString() };
});

// Root endpoint
fastify.get('/', async () => {
  return {
    name: '{{.AppName}}',
    description: '{{.Description}}',
    version: '0.1.0',
  };
});

// Example API endpoint with validation
const CreateItemSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
});

fastify.post('/api/items', async (request, reply) => {
  try {
    const data = CreateItemSchema.parse(request.body);

    {{if .DatabaseURL}}
    // Example database insert
    const [item] = await sql`
      INSERT INTO items (name, description)
      VALUES (${data.name}, ${data.description})
      RETURNING *
    `;

    return { success: true, item };
    {{else}}
    // No database configured, just echo back
    return { success: true, data };
    {{end}}
  } catch (err) {
    if (err instanceof z.ZodError) {
      reply.code(400);
      return { error: 'Validation failed', details: err.errors };
    }
    throw err;
  }
});

// Start server
const start = async () => {
  try {
    const port = parseInt(process.env.PORT || '3000', 10);
    const host = process.env.HOST || '0.0.0.0';

    await fastify.listen({ port, host });
    fastify.log.info(`Server listening on http://${host}:${port}`);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
};

start();
