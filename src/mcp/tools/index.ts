import { createDatabaseFactory } from "./createDatabase.js";
import { createWebAppFactory } from "./createWebApp.js";
import { openAppFactory } from "./openApp.js";
import { setupAppSchemaFactory } from "./setupAppSchema.js";
import { setupTestingFactory } from "./setupTesting.js";
import { uploadEnvToVercelFactory } from "./uploadEnvToVercel.js";
import { getViewSkillFactory } from "./viewSkill.js";
import { writeClaudeMdFactory } from "./writeClaudeMd.js";

export async function getApiFactories() {
  const viewSkillFactory = await getViewSkillFactory();

  return [
    createDatabaseFactory,
    createWebAppFactory,
    openAppFactory,
    setupAppSchemaFactory,
    setupTestingFactory,
    uploadEnvToVercelFactory,
    viewSkillFactory,
    writeClaudeMdFactory,
  ] as const;
}
