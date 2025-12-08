import { viewSkillFactory } from './viewSkill.js';
import { createDatabaseFactory } from './createDatabase.js';
import { createWebAppFactory } from './createWebApp.js';
import { openAppFactory } from './openApp.js';

export const apiFactories = [
  createDatabaseFactory,
  createWebAppFactory,
  openAppFactory,
  viewSkillFactory,
] as const;
