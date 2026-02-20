import type { HttpClient } from "./http.ts";

export interface TimeEntry {
  id: number;
  client_id: number;
  duration: number; // seconds
  started_at: string; // ISO datetime (UTC)
  local_started_at: string; // Local datetime (no timezone)
  note: string;
  billable: boolean;
}

export async function listTimeEntries(
  http: HttpClient,
  businessId: number,
  options: { startedFrom: string; startedTo: string }
): Promise<TimeEntry[]> {
  const entries = await http.getPaginated<TimeEntry>(
    `/timetracking/business/${businessId}/time_entries`,
    "time_entries",
    {
      started_from: `${options.startedFrom}T00:00:00`,
      started_to: `${options.startedTo}T23:59:59`,
    }
  );
  return entries;
}

export async function listUnbilledEntries(
  http: HttpClient,
  businessId: number,
  clientId: number
): Promise<TimeEntry[]> {
  const entries = await http.getPaginated<TimeEntry>(
    `/timetracking/business/${businessId}/time_entries`,
    "time_entries",
    {
      client_id: String(clientId),
      billed: "false",
      billable: "true",
    }
  );
  return entries;
}

export async function markEntriesAsBilled(
  http: HttpClient,
  businessId: number,
  entryIds: number[]
): Promise<void> {
  for (const id of entryIds) {
    await http.put(
      `/timetracking/business/${businessId}/time_entries/${id}`,
      { time_entry: { billed: true } }
    );
  }
}
