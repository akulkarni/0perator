import {
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { homedir, tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { addMCPServerViaJSON, expandPath } from "./mcpInstall.js";

describe("addMCPServerViaJSON", () => {
  let testDir: string;

  beforeEach(() => {
    // Create a unique temp directory for each test
    testDir = join(
      tmpdir(),
      `mcp-test-${Date.now()}-${Math.random().toString(36).slice(2)}`,
    );
    mkdirSync(testDir, { recursive: true });
  });

  afterEach(() => {
    // Clean up temp directory
    if (existsSync(testDir)) {
      rmSync(testDir, { recursive: true, force: true });
    }
  });

  it("should create a new config file with server entry", () => {
    const configPath = join(testDir, "mcp.json");

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);

    expect(config).toEqual({
      mcpServers: {
        tiger: {
          command: "tiger",
          args: ["mcp", "start"],
        },
      },
    });
  });

  it("should add server to existing config without mcpServers", () => {
    const configPath = join(testDir, "mcp.json");
    writeFileSync(
      configPath,
      JSON.stringify({ someOtherKey: "value" }, null, 2),
    );

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);

    expect(config).toEqual({
      someOtherKey: "value",
      mcpServers: {
        tiger: {
          command: "tiger",
          args: ["mcp", "start"],
        },
      },
    });
  });

  it("should add server to existing mcpServers without overwriting others", () => {
    const configPath = join(testDir, "mcp.json");
    const existingConfig = {
      mcpServers: {
        existing: {
          command: "existing-cmd",
          args: ["arg1"],
        },
      },
    };
    writeFileSync(configPath, JSON.stringify(existingConfig, null, 2));

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);

    expect(config).toEqual({
      mcpServers: {
        existing: {
          command: "existing-cmd",
          args: ["arg1"],
        },
        tiger: {
          command: "tiger",
          args: ["mcp", "start"],
        },
      },
    });
  });

  it("should overwrite existing server with same name", () => {
    const configPath = join(testDir, "mcp.json");
    const existingConfig = {
      mcpServers: {
        tiger: {
          command: "old-tiger",
          args: ["old-args"],
        },
      },
    };
    writeFileSync(configPath, JSON.stringify(existingConfig, null, 2));

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "new-tiger", [
      "new",
      "args",
    ]);

    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);

    expect(config.mcpServers.tiger).toEqual({
      command: "new-tiger",
      args: ["new", "args"],
    });
  });

  it("should preserve comments in JSON file", () => {
    const configPath = join(testDir, "mcp.json");
    const jsonWithComments = `{
  // This is a comment about mcpServers
  "mcpServers": {
    // Existing server comment
    "existing": {
      "command": "existing-cmd",
      "args": ["arg1"]
    }
  }
}`;
    writeFileSync(configPath, jsonWithComments);

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");

    // Check that comments are preserved
    expect(content).toContain("// This is a comment about mcpServers");
    expect(content).toContain("// Existing server comment");

    // Also verify the data is correct (parse strips comments for comparison)
    const config = JSON.parse(
      content.replace(/\/\/.*$/gm, "").replace(/\/\*[\s\S]*?\*\//g, ""),
    );
    expect(config.mcpServers.tiger).toEqual({
      command: "tiger",
      args: ["mcp", "start"],
    });
    expect(config.mcpServers.existing).toEqual({
      command: "existing-cmd",
      args: ["arg1"],
    });
  });

  it("should create nested directories if they don't exist", () => {
    const configPath = join(testDir, "deep", "nested", "dir", "mcp.json");

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    expect(existsSync(configPath)).toBe(true);
    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);
    expect(config.mcpServers.tiger).toBeDefined();
  });

  it("should handle empty config file", () => {
    const configPath = join(testDir, "mcp.json");
    writeFileSync(configPath, "");

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);

    expect(config.mcpServers.tiger).toEqual({
      command: "tiger",
      args: ["mcp", "start"],
    });
  });

  it("should handle whitespace-only config file", () => {
    const configPath = join(testDir, "mcp.json");
    writeFileSync(configPath, "   \n\n   ");

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");
    const config = JSON.parse(content);

    expect(config.mcpServers.tiger).toEqual({
      command: "tiger",
      args: ["mcp", "start"],
    });
  });

  it("should throw error for invalid JSON", () => {
    const configPath = join(testDir, "mcp.json");
    writeFileSync(configPath, "{ invalid json }");

    expect(() => {
      addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
        "mcp",
        "start",
      ]);
    }).toThrow(`Failed to parse existing config at ${configPath}`);
  });

  it("should handle trailing commas in JSON", () => {
    const configPath = join(testDir, "mcp.json");
    const jsonWithTrailingComma = `{
  "mcpServers": {
    "existing": {
      "command": "cmd",
      "args": ["arg"],
    },
  },
}`;
    writeFileSync(configPath, jsonWithTrailingComma);

    addMCPServerViaJSON(configPath, "/mcpServers", "tiger", "tiger", [
      "mcp",
      "start",
    ]);

    const content = readFileSync(configPath, "utf-8");
    // comment-json handles trailing commas
    expect(content).toContain('"tiger"');
  });
});

describe("expandPath", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    vi.resetModules();
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it("should expand ~ to home directory", () => {
    const result = expandPath("~/some/path");
    expect(result).toBe(join(homedir(), "some/path"));
  });

  it("should expand standalone ~", () => {
    const result = expandPath("~");
    expect(result).toBe(homedir());
  });

  it("should not expand ~ in the middle of path", () => {
    const result = expandPath("/some/~/path");
    expect(result).toBe("/some/~/path");
  });

  it("should expand $VAR style environment variables", () => {
    process.env.TEST_VAR = "/test/value";
    const result = expandPath("$TEST_VAR/subpath");
    expect(result).toBe("/test/value/subpath");
  });

  it("should expand curly brace style environment variables", () => {
    process.env.TEST_VAR = "/test/value";
    // biome-ignore lint/suspicious/noTemplateCurlyInString: testing actual ${VAR} expansion
    const result = expandPath("${TEST_VAR}/subpath");
    expect(result).toBe("/test/value/subpath");
  });

  it("should expand multiple environment variables", () => {
    process.env.VAR1 = "first";
    process.env.VAR2 = "second";
    const result = expandPath("$VAR1/$VAR2/end");
    expect(result).toBe("first/second/end");
  });

  it("should replace undefined env vars with empty string", () => {
    delete process.env.UNDEFINED_VAR;
    const result = expandPath("$UNDEFINED_VAR/path");
    expect(result).toBe("/path");
  });

  it("should handle both ~ and env vars together", () => {
    process.env.SUBDIR = "mydir";
    const result = expandPath("~/$SUBDIR/file");
    expect(result).toBe(join(homedir(), "mydir/file"));
  });

  it("should return path unchanged if no expansions needed", () => {
    const result = expandPath("/absolute/path/to/file");
    expect(result).toBe("/absolute/path/to/file");
  });

  it("should handle empty string", () => {
    const result = expandPath("");
    expect(result).toBe("");
  });
});
