import { createDatabaseFactory } from "./createDatabase.js";
import { createWebAppFactory } from "./createWebApp.js";
import { openAppFactory } from "./openApp.js";
import { uploadEnvToVercelFactory } from "./uploadEnvToVercel.js";
import { getViewSkillFactory } from "./viewSkill.js";

export async function getApiFactories() {
  const viewSkillFactory = await getViewSkillFactory();

  return [
    createDatabaseFactory,
    createWebAppFactory,
    openAppFactory,
    uploadEnvToVercelFactory,
    viewSkillFactory,
  ] as const;
}
