import { mkdir, readdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import Handlebars from "handlebars";
import { templatesDir } from "../config.js";

export interface AppTemplateVars {
  app_name: string;
  use_auth: boolean;
  product_brief?: string | undefined;
  future_features?: string | undefined;
}

type ContentTransform = (content: string) => string;

/**
 * Copy a template directory to destination, optionally transforming file contents
 */
async function copyTemplateDir(
  templateName: string,
  destDir: string,
  transform?: ContentTransform,
): Promise<void> {
  const srcBaseDir = join(templatesDir, templateName);

  async function copyDir(srcDir: string): Promise<void> {
    const entries = await readdir(srcDir, { withFileTypes: true });

    for (const entry of entries) {
      const srcPath = join(srcDir, entry.name);
      const relPath = relative(srcBaseDir, srcPath);
      const destPath = join(destDir, relPath);

      if (entry.isDirectory()) {
        await mkdir(destPath, { recursive: true });
        await copyDir(srcPath);
      } else {
        await mkdir(dirname(destPath), { recursive: true });

        const content = await readFile(srcPath, "utf-8");
        const output = transform ? transform(content) : content;
        await writeFile(destPath, output);
      }
    }
  }

  await copyDir(srcBaseDir);
}

/**
 * Write app templates with Handlebars templating
 */
export async function writeAppTemplates(
  destDir: string,
  vars: AppTemplateVars,
): Promise<void> {
  await copyTemplateDir("app", destDir, (content) => {
    const template = Handlebars.compile(content);
    return template(vars);
  });
}

/**
 * Write testing templates (static files, no templating)
 */
export async function writeTestingTemplates(destDir: string): Promise<void> {
  await copyTemplateDir("testing", destDir);
}
