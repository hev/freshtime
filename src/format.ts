export interface WeeklySummary {
  weekStart: string; // YYYY-MM-DD
  weekEnd: string;   // YYYY-MM-DD
  clients: ClientSummary[];
  grandTotal: number;
}

export interface ClientSummary {
  name: string;
  daily: number[]; // hours for Mon–Fri (5 elements)
  total: number;
}

const DAY_HEADERS = ["Mon", "Tue", "Wed", "Thu", "Fri"];

function formatHours(h: number): string {
  if (h === 0) return "—";
  return h.toFixed(1);
}

function formatDateRange(start: string, end: string): string {
  const s = new Date(start + "T00:00:00");
  const e = new Date(end + "T00:00:00");
  const months = [
    "Jan", "Feb", "Mar", "Apr", "May", "Jun",
    "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
  ];
  return `${months[s.getMonth()]} ${s.getDate()} – ${months[e.getMonth()]} ${e.getDate()}, ${e.getFullYear()}`;
}

export function formatTable(summary: WeeklySummary): string {
  const COL_WIDTH = 6;
  const NAME_WIDTH = 20;
  const lines: string[] = [];

  lines.push(`Week of ${formatDateRange(summary.weekStart, summary.weekEnd)}`);
  lines.push("");

  // Header
  const header =
    "Client".padEnd(NAME_WIDTH) +
    DAY_HEADERS.map((d) => d.padStart(COL_WIDTH)).join("") +
    "  Total";
  lines.push(header);

  const separator = "─".repeat(header.length);
  lines.push(separator);

  // Client rows
  for (const client of summary.clients) {
    const row =
      client.name.slice(0, NAME_WIDTH).padEnd(NAME_WIDTH) +
      client.daily.map((h) => formatHours(h).padStart(COL_WIDTH)).join("") +
      formatHours(client.total).padStart(COL_WIDTH + 1) +
      "h";
    lines.push(row);
  }

  lines.push(separator);

  // Totals row
  const dailyTotals = [0, 0, 0, 0, 0];
  for (const client of summary.clients) {
    for (let i = 0; i < 5; i++) {
      dailyTotals[i]! += client.daily[i]!;
    }
  }
  const totalsRow =
    "Total".padEnd(NAME_WIDTH) +
    dailyTotals.map((h) => formatHours(h).padStart(COL_WIDTH)).join("") +
    formatHours(summary.grandTotal).padStart(COL_WIDTH + 1) +
    "h";
  lines.push(totalsRow);

  return lines.join("\n");
}

export function formatJson(summary: WeeklySummary): string {
  return JSON.stringify(summary, null, 2);
}
