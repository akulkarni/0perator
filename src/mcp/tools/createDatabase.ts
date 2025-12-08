import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { exec } from 'child_process';
import { promisify } from 'util';
import { ServerContext } from '../../types.js';

const execAsync = promisify(exec);

const inputSchema = {
  name: z
    .string()
    .optional()
    .describe('Database name (default: app-db)'),
} as const;

const outputSchema = {
  success: z.boolean().describe('Whether the database was created successfully'),
  service_id: z.string().optional().describe('The Tiger Cloud service ID'),
  error: z.string().optional().describe('Error message if creation failed'),
} as const;

type OutputSchema = {
  success: boolean;
  service_id?: string;
  error?: string;
};

export const createDatabaseFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'create_database',
    config: {
      title: 'Create Database',
      description:
        '🗄️ Set up any database - PostgreSQL on Tiger Cloud (default, FREE). Auto-configures with schema, migrations, and connection handling. Use for any database request.',
      inputSchema,
      outputSchema,
    },
    fn: async ({ name }): Promise<OutputSchema> => {
      const dbName = name || 'app-db';

      const cmdArgs = [
        'tiger',
        'service',
        'create',
        '--name',
        dbName,
        '--cpu',
        'shared',
        '--memory',
        'shared',
        '--addons',
        'time-series,ai',
        '--wait-timeout',
        '2m',
        '-o',
        'json',
      ];

      try {
        const { stdout, stderr } = await execAsync(cmdArgs.join(' '));
        const result = JSON.parse(stdout) as { service_id?: string };

        if (!result.service_id) {
          return {
            success: false,
            error: `No service_id in response: ${stdout}${stderr}`,
          };
        }

        return {
          success: true,
          service_id: result.service_id,
        };
      } catch (err) {
        const error = err as Error & { stdout?: string; stderr?: string };
        return {
          success: false,
          error: `Failed to create database: ${error.message}\n${error.stdout || ''}${error.stderr || ''}`,
        };
      }
    },
  };
};
