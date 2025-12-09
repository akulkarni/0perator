import { checkbox } from "@inquirer/prompts";
import { Command } from "commander";
import pc from "picocolors";
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

      // If no clients specified, prompt interactively
      if (clients.length === 0) {
        console.log(pc.bold("\n0perator Setup\n"));
        clients = await checkbox({
          message:
            "Select IDEs to configure (space to select, enter to confirm)",
          theme: {
            prefix: {
              idle: pc.cyan("›"),
              done: pc.green("✓"),
            },
          },
          choices: supportedClients.map((c) => ({
            name: c.displayName,
            value: c.name,
          })),
        });
      }

      if (clients.length === 0) {
        console.log(pc.yellow("No IDEs selected. Exiting."));
        return;
      }

      console.log(pc.blue("\nConfiguring MCP servers...\n"));

      let successCount = 0;
      let failCount = 0;

      for (const clientName of clients) {
        const client = supportedClients.find((c) => c.name === clientName);
        if (!client) {
          console.log(pc.red(`Unknown client: ${clientName}`));
          failCount++;
          continue;
        }

        try {
          console.log(`  ${pc.cyan("→")} ${client.displayName}...`);
          await installBoth(clientName, { devMode: options.dev });
          console.log(`  ${pc.green("✓")} ${client.displayName} configured`);
          successCount++;
        } catch (err) {
          const error = err as Error;
          console.log(
            `  ${pc.red("✗")} ${client.displayName}: ${error.message}`,
          );
          failCount++;
        }
      }

      if (failCount === 0) {
        console.log(
          pc.green("\nDone! Restart your IDE to use the MCP servers."),
        );
      } else if (successCount === 0) {
        console.log(pc.red("\nFailed to configure any IDEs."));
      } else {
        console.log(
          pc.yellow(
            `\nPartially completed: ${successCount} succeeded, ${failCount} failed.`,
          ),
        );
      }
    });

  return init;
}

function collect(value: string, previous: string[]): string[] {
  return previous.concat([value]);
}
