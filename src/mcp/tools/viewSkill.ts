import { ApiFactory } from '@tigerdata/mcp-boilerplate';
import { z } from 'zod';
import { ServerContext } from '../../types.js';
import { skills, viewSkillContent } from '../skillutils/index.js';

// Create enum schema dynamically from loaded skills
const inputSchema = {
  name: z
    .enum(Array.from(skills.keys()) as [string, ...string[]])
    .describe('Skill name (directory name)'),
} as const;

const outputSchema = {
  success: z.boolean(),
  name: z.string(),
  description: z.string(),
  body: z.string(),
} as const;

type OutputSchema = {
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
};

export const viewSkillFactory: ApiFactory<
  ServerContext,
  typeof inputSchema,
  typeof outputSchema
> = () => {
  return {
    name: 'view_skill',
    config: {
      title: 'View Skill',
      description: `📖 View instructions for a specific skill by name.

Available skills:
${Array.from(skills.values())
  .map((s) => `- ${s.name}: ${s.description}`)
  .join('\n')}
`,
      inputSchema,
      outputSchema,
    },
    fn: async ({ name }): Promise<OutputSchema> => {
      const skill = skills.get(name);

      if (!skill) {
        throw new Error(`Skill '${name}' not found`);
      }

      const body = await viewSkillContent(name);

      return {
        success: true,
        name: skill.name,
        description: skill.description || '',
        body,
      };
    },
  };
};
