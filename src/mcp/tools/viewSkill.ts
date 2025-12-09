import type { ApiFactory } from "@tigerdata/mcp-boilerplate";
import { z } from "zod";
import type { ServerContext } from "../../types.js";
import type { Skill } from "../skillutils/index.js";
import { loadSkills, viewSkillContent } from "../skillutils/index.js";

const outputSchema = {
  success: z.boolean(),
  name: z.string(),
  description: z.string(),
  body: z.string(),
} as const;

type OutputSchema = {
  [K in keyof typeof outputSchema]: z.infer<(typeof outputSchema)[K]>;
};

export function createViewSkillFactory(
  skills: Map<string, Skill>,
): ApiFactory<
  ServerContext,
  { name: z.ZodEnum<[string, ...string[]]> },
  typeof outputSchema
> {
  const inputSchema = {
    name: z
      .enum(Array.from(skills.keys()) as [string, ...string[]])
      .describe("Skill name (directory name)"),
  } as const;

  return () => ({
    name: "view_skill",
    config: {
      title: "View Skill",
      description: `📖 View instructions for a specific skill by name.

Available skills:
${Array.from(skills.values())
  .map((s) => `- ${s.name}: ${s.description}`)
  .join("\n")}
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
        description: skill.description || "",
        body,
      };
    },
  });
}

// Helper to get the factory with loaded skills
export async function getViewSkillFactory() {
  const skills = await loadSkills();
  return createViewSkillFactory(skills);
}
