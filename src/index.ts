#!/usr/bin/env node
import { Command } from 'commander';
import { version } from './config.js';
import { createMcpCommand } from './commands/mcp.js';

const program = new Command();

program
  .name('0perator')
  .description('Infrastructure for AI native development')
  .version(version);

program.addCommand(createMcpCommand());

program.parse();
