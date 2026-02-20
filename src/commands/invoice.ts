import type { HttpClient } from "../api/http.ts";
import type { Config } from "../config.ts";
import {
  listUnbilledEntries,
  markEntriesAsBilled,
  type TimeEntry,
} from "../api/time-entries.ts";
import {
  createInvoice,
  type InvoiceLine,
  type CreateInvoiceRequest,
} from "../api/invoices.ts";

export interface InvoiceOptions {
  rate?: string;
  currency?: string;
  dryRun?: boolean;
}

export function buildInvoiceLines(
  entries: TimeEntry[],
  rate: string,
  currency: string
): InvoiceLine[] {
  return entries.map((entry) => ({
    type: 0 as const,
    name: entry.note || "Consulting",
    description: entry.local_started_at.split("T")[0]!,
    qty: (entry.duration / 3600).toFixed(2),
    unit_cost: { amount: rate, code: currency },
  }));
}

export async function runInvoice(
  http: HttpClient,
  config: Config,
  clientId: number,
  options: InvoiceOptions
): Promise<string> {
  const entries = await listUnbilledEntries(http, config.business_id, clientId);

  if (entries.length === 0) {
    return "No unbilled time entries found for this client.";
  }

  const rate =
    options.rate ??
    config.client_rates?.[String(clientId)];

  if (!rate) {
    throw new Error(
      `No rate configured for client ${clientId}. ` +
        `Use --rate <amount> or set client_rates.${clientId} in config.`
    );
  }

  const currency = options.currency ?? config.default_currency ?? "USD";
  const lines = buildInvoiceLines(entries, rate, currency);

  const totalHours = entries.reduce((sum, e) => sum + e.duration, 0) / 3600;
  const totalAmount = (totalHours * parseFloat(rate)).toFixed(2);

  if (options.dryRun) {
    const output: string[] = [];
    output.push("Dry run — no invoice created.\n");
    output.push(`Entries:  ${entries.length}`);
    output.push(`Hours:   ${totalHours.toFixed(2)}`);
    output.push(`Rate:    ${rate} ${currency}/hr`);
    output.push(`Total:   ${totalAmount} ${currency}\n`);
    output.push("Line items:");
    for (const line of lines) {
      output.push(`  ${line.description}  ${line.qty}h  ${line.name}`);
    }
    return output.join("\n");
  }

  const today = new Date().toISOString().split("T")[0]!;
  const request: CreateInvoiceRequest = {
    invoice: {
      customerid: clientId,
      create_date: today,
      lines,
      status: 2,
    },
  };

  const invoice = await createInvoice(http, config.account_id, request);
  await markEntriesAsBilled(
    http,
    config.business_id,
    entries.map((e) => e.id)
  );

  const output: string[] = [];
  output.push(`Invoice #${invoice.invoice_number} created.`);
  output.push(`Entries:  ${entries.length}`);
  output.push(`Hours:   ${totalHours.toFixed(2)}`);
  output.push(`Total:   ${invoice.amount.amount} ${invoice.amount.code}`);
  output.push(`Link:    ${invoice.links.client_view}`);
  return output.join("\n");
}
