import { viewSkillFactory } from './viewSkill.js';
import { createDatabaseFactory } from './createDatabase.js';
import { createWebAppFactory } from './createWebApp.js';

export const apiFactories = [
  createDatabaseFactory,
  createWebAppFactory,
  viewSkillFactory,
] as const;
