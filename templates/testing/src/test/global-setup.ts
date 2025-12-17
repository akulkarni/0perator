import { execSync } from "child_process";
import { config } from "dotenv";

export default function setup() {
  // globalSetup runs in a separate process before Vitest loads env files,
  // so we manually load .env.test.local here for drizzle-kit to use
  config({ path: ".env.test.local" });

  // Push schema to test database before running tests
  // This ensures the test schema matches your Drizzle schema
  execSync("npx drizzle-kit push", { stdio: "inherit" });
}
