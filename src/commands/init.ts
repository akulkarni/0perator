import { existsSync } from "node:fs";
import { join } from "node:path";
import pc from "picocolors";
import * as p from "@clack/prompts";
import { Command } from "commander";
import { packageRoot } from "../config.js";
import { supportedClients } from "../lib/clients.js";
import { installBoth } from "../lib/install.js";

interface InitOptions {
  client?: string;
  dev?: boolean;
  latest?: boolean;
}

function printBanner(): void {
  const accent = pc.cyan;
  console.log();
  console.log(accent("     ██████╗ ██████╗ ███████╗██████╗  █████╗ ████████╗ ██████╗ ██████╗ "));
  console.log(accent("    ██╔═████╗██╔══██╗██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗"));
  console.log(accent("    ██║██╔██║██████╔╝█████╗  ██████╔╝███████║   ██║   ██║   ██║██████╔╝"));
  console.log(accent("    ████╔╝██║██╔═══╝ ██╔══╝  ██╔══██╗██╔══██║   ██║   ██║   ██║██╔══██╗"));
  console.log(accent("    ╚██████╔╝██║     ███████╗██║  ██║██║  ██║   ██║   ╚██████╔╝██║  ██║"));
  console.log(accent("     ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝"));
  console.log();
  console.log(accent("               Infrastructure for AI native development"));
  console.log();
  console.log("──────────────────────────────────────────────────────────────────────────");
}

export function createInitCommand(): Command {
  const init = new Command("init")
    .description("Configure IDEs with MCP servers")
    .option("--client <name>", "Client to configure")
    .option("--dev", "Use development mode")
    .option("--no-latest", "Pin to current version instead of using latest")
    .action(async (options: InitOptions) => {
      // Check if --dev is used outside a development context
      if (options.dev) {
        const gitDir = join(packageRoot, ".git");
        if (!existsSync(gitDir)) {
          console.error(
            "Error: --dev flag can only be used when running from a local git checkout of 0perator.",
          );
          console.error(
            "For development, clone the repo and run: npm run dev -- init --dev",
          );
          process.exit(1);
        }
      }

      printBanner();

      let clientName = options.client;

      // If no client specified, prompt interactively
      if (!clientName) {
        const selected = await p.select({
          message: "Select IDE to configure",
          options: supportedClients.map((c) => ({
            label: c.displayName,
            value: c.name,
          })),
        });

        if (p.isCancel(selected)) {
          p.cancel("Setup cancelled.");
          process.exit(0);
        }

        clientName = selected as string;
      }

      const client = supportedClients.find((c) => c.name === clientName);
      if (!client) {
        p.log.error(`Unknown client: ${clientName}`);
        process.exit(1);
      }

      const s = p.spinner();
      s.start(`Configuring ${client.displayName}...`);

      try {
        await installBoth(clientName, { devMode: options.dev, latest: options.latest });
        s.stop(`${client.displayName} configured`);
        p.outro("Done! Restart your IDE to use the MCP servers.");
        console.log("");
        console.log("Try asking your AI coding assistant:");
        console.log("  • Create a new collaborative TODO webapp");
        console.log("  • Build a real-time chat application");
        console.log("  • Create a dashboard to track my fitness goals");
        console.log("");
      } catch (err) {
        const error = err as Error;
        s.stop(`${client.displayName} failed`);
        p.log.error(error.message);
        process.exit(1);
      }
    });

  return init;
}
