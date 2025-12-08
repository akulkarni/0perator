import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));

// Package root directory (relative to dist/)
export const packageRoot = join(__dirname, "..");

// Skills directory at package root level
export const skillsDir = join(packageRoot, "skills");

// Templates directory at package root level
export const templatesDir = join(packageRoot, "templates");

// Read version from package.json
const pkg = JSON.parse(
  readFileSync(join(packageRoot, "package.json"), "utf-8"),
);
export const version: string = pkg.version;
