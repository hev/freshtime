import { homedir } from "os";
import { join } from "path";
import { mkdir, readFile, writeFile } from "fs/promises";

export interface Config {
  access_token: string;
  refresh_token?: string;
  account_id: string;
  business_id: number;
  client_rates?: Record<string, string>; // client_id -> hourly rate (e.g. "150.00")
  default_currency?: string;             // e.g. "USD"
}

const CONFIG_DIR = join(homedir(), ".config", "freshtime");
const CONFIG_PATH = join(CONFIG_DIR, "config.json");

export function getConfigPath(): string {
  return CONFIG_PATH;
}

export async function loadConfig(): Promise<Config> {
  try {
    const raw = await readFile(CONFIG_PATH, "utf-8");
    return JSON.parse(raw) as Config;
  } catch {
    throw new Error(
      "Config not found. Run `freshtime setup` to configure your token."
    );
  }
}

export async function saveConfig(config: Config): Promise<void> {
  await mkdir(CONFIG_DIR, { recursive: true });
  await writeFile(CONFIG_PATH, JSON.stringify(config, null, 2) + "\n");
}
