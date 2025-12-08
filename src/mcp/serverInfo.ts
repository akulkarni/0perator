import { ServerContext } from '../types.js';
import { version } from '../config.js';

export const serverInfo = {
  name: '0perator',
  version,
} as const;

export const context: ServerContext = {};
