import { createDatabaseFactory } from "./createDatabase.js";
import { createWebAppFactory } from "./createWebApp.js";
import { openAppFactory } from "./openApp.js";
import { getViewSkillFactory } from "./viewSkill.js";

export async function getApiFactories() {
  const viewSkillFactory = await getViewSkillFactory();

  return [
    createDatabaseFactory,
    createWebAppFactory,
    openAppFactory,
    viewSkillFactory,
  ] as const;
}
