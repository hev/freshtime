import { loadConfig, saveConfig } from "../config.ts";
import { refreshAccessToken } from "../api/http.ts";

export async function runRefresh(): Promise<void> {
  const config = await loadConfig();

  if (!config.refresh_token) {
    throw new Error("No refresh token in config. Run `freshtime setup` first.");
  }

  const tokens = await refreshAccessToken(config.refresh_token);

  await saveConfig({
    ...config,
    access_token: tokens.access_token,
    refresh_token: tokens.refresh_token,
  });
}
