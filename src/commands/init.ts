import * as p from "@clack/prompts";
import { Command } from "commander";
import { supportedClients } from "../lib/clients.js";
import { installBoth } from "../lib/install.js";

interface InitOptions {
  client?: string[];
  dev?: boolean;
}

export function createInitCommand(): Command {
  const init = new Command("init")
    .description("Configure IDEs with MCP servers")
    .option(
      "--client <name>",
      "Client to configure (can be repeated)",
      collect,
      [],
    )
    .option("--dev", "Use development mode")
    .action(async (options: InitOptions) => {
      let clients = options.client || [];

      p.intro("0perator Setup");

      // If no clients specified, prompt interactively
      if (clients.length === 0) {
        const selected = await p.multiselect({
          message: "Select IDEs to configure",
          options: supportedClients.map((c) => ({
            label: c.displayName,
            value: c.name,
          })),
          required: false,
        });

        if (p.isCancel(selected)) {
          p.cancel("Setup cancelled.");
          process.exit(0);
        }

        clients = selected as string[];
      }

      if (clients.length === 0) {
        p.outro("No IDEs selected.");
        return;
      }

      const s = p.spinner();

      let successCount = 0;
      let failCount = 0;

      for (const clientName of clients) {
        const client = supportedClients.find((c) => c.name === clientName);
        if (!client) {
          p.log.error(`Unknown client: ${clientName}`);
          failCount++;
          continue;
        }

        s.start(`Configuring ${client.displayName}...`);

        try {
          await installBoth(clientName, { devMode: options.dev });
          s.stop(`${client.displayName} configured`);
          successCount++;
        } catch (err) {
          const error = err as Error;
          s.stop(`${client.displayName} failed`);
          p.log.error(error.message);
          failCount++;
        }
      }

      if (failCount === 0) {
        p.outro("Done! Restart your IDE to use the MCP servers.");
      } else if (successCount === 0) {
        p.outro("Failed to configure any IDEs.");
      } else {
        p.outro(
          `Partially completed: ${successCount} succeeded, ${failCount} failed.`,
        );
      }
    });

  return init;
}

function collect(value: string, previous: string[]): string[] {
  return previous.concat([value]);
}
