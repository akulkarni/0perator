import { mkdir, readdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import { templatesDir } from "../config.js";

/**
 * Copy app templates (CLAUDE.md, globals.css, etc.) to destination
 */
export async function writeAppTemplates(destDir: string): Promise<void> {
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
        const content = await readFile(srcPath);
        await writeFile(destPath, content);
      }
    }
  }

  await copyDir(appDir, destDir);
}
