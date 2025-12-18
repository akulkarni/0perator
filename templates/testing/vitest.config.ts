import path from "node:path";
import { defineConfig } from "vitest/config";
import { loadEnv } from "vite";

export default defineConfig({
  test: {
    environment: "node",
    globals: true,
    env: loadEnv("test.local", process.cwd(), ""), // loads .env.test.local
    globalSetup: ["./src/test/global-setup.ts"],
    //the following two settings prevent tests for interfering with each other on same db
    fileParallelism: false,
    sequence: {
      concurrent: false
    }
  },
  resolve: {
    alias: {
      "~": path.resolve(__dirname, "./src"),
    },
  },
});
