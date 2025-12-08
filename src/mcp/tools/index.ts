import { viewSkillFactory } from './viewSkill.js';
import { createDatabaseFactory } from './createDatabase.js';

export const apiFactories = [
  createDatabaseFactory,
  viewSkillFactory,
] as const;
