#!/usr/bin/env node
import { Command } from "commander";
import { createInitCommand } from "./commands/init.js";
import { createMcpCommand } from "./commands/mcp.js";
import { version } from "./config.js";

const program = new Command();

program
  .name("0perator")
  .description("Infrastructure for AI native development")
  .version(version);

program.addCommand(createInitCommand());
program.addCommand(createMcpCommand());

program.parse();
