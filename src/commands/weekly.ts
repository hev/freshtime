import type { HttpClient } from "../api/http.ts";
import { listTimeEntries, type TimeEntry } from "../api/time-entries.ts";
import { listClients } from "../api/clients.ts";
import type { WeeklySummary, ClientSummary } from "../format.ts";

export function getWeekRange(referenceDate: Date): {
  weekStart: string;
  weekEnd: string;
} {
  const d = new Date(referenceDate);
  // Get Monday: getDay() returns 0=Sun, 1=Mon...
  const day = d.getDay();
  const diffToMonday = day === 0 ? -6 : 1 - day;
  const monday = new Date(d);
  monday.setDate(d.getDate() + diffToMonday);

  const friday = new Date(monday);
  friday.setDate(monday.getDate() + 4);

  const fmt = (dt: Date) =>
    dt.getFullYear() +
    "-" +
    String(dt.getMonth() + 1).padStart(2, "0") +
    "-" +
    String(dt.getDate()).padStart(2, "0");

  return { weekStart: fmt(monday), weekEnd: fmt(friday) };
}

export function buildSummary(
  entries: TimeEntry[],
  clientNames: Map<number, string>,
  weekStart: string
): WeeklySummary {
  const monday = new Date(weekStart + "T00:00:00");
  const friday = new Date(monday);
  friday.setDate(monday.getDate() + 4);

  const weekEnd =
    friday.getFullYear() +
    "-" +
    String(friday.getMonth() + 1).padStart(2, "0") +
    "-" +
    String(friday.getDate()).padStart(2, "0");

  // Group by client
  const byClient = new Map<number, number[]>();

  for (const entry of entries) {
    // Use local_started_at (no timezone) to determine day-of-week
    const localDate = entry.local_started_at ?? entry.started_at;
    const entryDate = new Date(localDate);
    // Determine day index (0=Mon, 4=Fri)
    const dayOfWeek = entryDate.getDay();
    const dayIndex = dayOfWeek === 0 ? 6 : dayOfWeek - 1; // 0=Mon...6=Sun

    if (dayIndex < 0 || dayIndex > 4) continue; // Skip weekends

    if (!byClient.has(entry.client_id)) {
      byClient.set(entry.client_id, [0, 0, 0, 0, 0]);
    }
    const daily = byClient.get(entry.client_id)!;
    daily[dayIndex]! += entry.duration;
  }

  // Build client summaries
  const clients: ClientSummary[] = [];
  let grandTotal = 0;

  for (const [clientId, dailySeconds] of byClient) {
    const dailyHours = dailySeconds.map((s) => +(s / 3600).toFixed(2));
    const total = +(dailyHours.reduce((a, b) => a + b, 0)).toFixed(2);
    grandTotal += total;

    clients.push({
      name: clientNames.get(clientId) ?? `Client #${clientId}`,
      daily: dailyHours,
      total,
    });
  }

  // Sort by name
  clients.sort((a, b) => a.name.localeCompare(b.name));
  grandTotal = +(grandTotal.toFixed(2));

  return { weekStart, weekEnd, clients, grandTotal };
}

export async function runWeekly(
  http: HttpClient,
  accountId: string,
  businessId: number,
  options: { weekOf?: string }
): Promise<WeeklySummary> {
  const ref = options.weekOf ? new Date(options.weekOf + "T00:00:00") : new Date();
  const { weekStart, weekEnd } = getWeekRange(ref);

  const [entries, clientNames] = await Promise.all([
    listTimeEntries(http, businessId, {
      startedFrom: weekStart,
      startedTo: weekEnd,
    }),
    listClients(http, accountId),
  ]);

  return buildSummary(entries, clientNames, weekStart);
}
