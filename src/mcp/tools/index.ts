import { createViewSkillToolFactory } from "@tigerdata/mcp-boilerplate/skills";
import { createDatabaseFactory } from "./createDatabase.js";
import { createWebAppFactory } from "./createWebApp.js";
import { openAppFactory } from "./openApp.js";
import { setupAppSchemaFactory } from "./setupAppSchema.js";
import { setupTestingFactory } from "./setupTesting.js";
import { uploadEnvToVercelFactory } from "./uploadEnvToVercel.js";
import { writeClaudeMdFactory } from "./writeClaudeMd.js";

export const apiFactories = [
  createDatabaseFactory,
  createWebAppFactory,
  openAppFactory,
  setupAppSchemaFactory,
  setupTestingFactory,
  uploadEnvToVercelFactory,
  writeClaudeMdFactory,
  createViewSkillToolFactory({
    appendSkillsListToDescription: true,
    name: "view_skill",
  }),
] as const;
