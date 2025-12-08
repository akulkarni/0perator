import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { exec } from 'child_process';
import { ServerContext } from '../../types.js';

const inputSchema = {
  url: z
    .string()
    .optional()
    .describe('URL to open (default: http://localhost:3000)'),
} as const;

const outputSchema = {
  success: z.boolean().describe('Whether the browser was opened successfully'),
  message: z.string().describe('Status message'),
  url: z.string().describe('The URL that was opened'),
} as const;

type OutputSchema = {
  success: boolean;
  message: string;
  url: string;
};

export const openAppFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'open_app',
    config: {
      title: 'Open App',
      description:
        '🌐 Open the app in browser. Call this AFTER all setup (database, auth, UI) is complete to show the user their running app.',
      inputSchema,
      outputSchema,
    },
    fn: async ({ url }): Promise<OutputSchema> => {
      const targetUrl = url || 'http://localhost:3000';

      return new Promise((resolve) => {
        // Use 'open' command on macOS
        exec(`open "${targetUrl}"`, (err) => {
          if (err) {
            resolve({
              success: false,
              message: `Failed to open browser: ${err.message}`,
              url: targetUrl,
            });
          } else {
            resolve({
              success: true,
              message: `Opened ${targetUrl} in browser`,
              url: targetUrl,
            });
          }
        });
      });
    },
  };
};
