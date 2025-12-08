import { version } from "../config.js";
import type { ServerContext } from "../types.js";

export const serverInfo = {
  name: "0perator",
  version,
} as const;

export const context: ServerContext = {};
