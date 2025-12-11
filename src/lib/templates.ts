import { mkdir, readdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import Handlebars from "handlebars";
import { templatesDir } from "../config.js";

export interface TemplateVars {
  app_name: string;
  use_auth?: boolean;
  product_brief?: string;
  future_features?: string;
}

/**
 * Copy app templates (CLAUDE.md, globals.css, etc.) to destination
 * Applies Handlebars templating to all files
 */
export async function writeAppTemplates(destDir: string, vars: TemplateVars): Promise<void> {
  const appDir = join(templatesDir, "app");

  async function copyDir(srcDir: string, destBase: string): Promise<void> {
    const entries = await readdir(srcDir, { withFileTypes: true });

    for (const entry of entries) {
      const srcPath = join(srcDir, entry.name);
      const relPath = relative(appDir, srcPath);
      const destPath = join(destBase, relPath);

      if (entry.isDirectory()) {
        await mkdir(destPath, { recursive: true });
        await copyDir(srcPath, destBase);
      } else {
        // Ensure parent directory exists
        await mkdir(dirname(destPath), { recursive: true });

        const content = await readFile(srcPath, "utf-8");
        const template = Handlebars.compile(content);
        const rendered = template(vars);

        await writeFile(destPath, rendered);
      }
    }
  }

  await copyDir(appDir, destDir);
}
