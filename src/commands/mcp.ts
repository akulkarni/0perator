import { Command } from "commander";
import { startMcpServer } from "../mcp/server.js";

export interface McpStartOptions {
  disableViewSkill?: boolean;
}

export function createMcpCommand(): Command {
  const mcp = new Command("mcp").description("MCP server commands");

  mcp
    .command("start")
    .description("Start the MCP server")
    .option(
      "--disable-view-skill",
      "Disable the view_skill tool (use when skills are provided natively by a plugin)"
    )
    .action(async (options: McpStartOptions) => {
      const serverOptions: import("../mcp/server.js").McpServerOptions = {};
      if (options.disableViewSkill) {
        serverOptions.disableViewSkill = true;
      }
      await startMcpServer(serverOptions);
    });

  return mcp;
}
