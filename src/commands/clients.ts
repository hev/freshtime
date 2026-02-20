import type { HttpClient } from "../api/http.ts";
import { listClients } from "../api/clients.ts";

export async function runClients(
  http: HttpClient,
  accountId: string
): Promise<string> {
  const clients = await listClients(http, accountId);

  const lines: string[] = [];
  const idWidth = 8;
  lines.push("ID".padEnd(idWidth) + "Name");
  lines.push("─".repeat(40));

  for (const [id, name] of clients) {
    lines.push(String(id).padEnd(idWidth) + name);
  }

  if (clients.size === 0) {
    lines.push("No clients found.");
  }

  return lines.join("\n");
}
