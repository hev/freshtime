#!/usr/bin/env bun
import { Command } from "commander";
import { loadConfig } from "./config.ts";
import { createHttpClient } from "./api/http.ts";
import { AuthError } from "./api/http.ts";
import { runSetup } from "./commands/setup.ts";
import { runWeekly } from "./commands/weekly.ts";
import { runClients } from "./commands/clients.ts";
import { runInvoice } from "./commands/invoice.ts";
import { runRefresh } from "./commands/refresh.ts";
import { formatTable, formatJson } from "./format.ts";

const program = new Command();

program
  .name("freshtime")
  .description("FreshBooks weekly time summary CLI")
  .version("1.0.0");

program
  .command("setup")
  .description("Configure your FreshBooks access token")
  .action(async () => {
    try {
      await runSetup();
    } catch (err) {
      if (err instanceof AuthError) {
        console.error("Error: Invalid token. Please check your access token.");
      } else {
        console.error("Error:", (err as Error).message);
      }
      process.exit(1);
    }
  });

program
  .command("weekly")
  .description("Show weekly time summary grouped by client")
  .option("--week-of <date>", "Show week containing this date (YYYY-MM-DD)")
  .option("--json", "Output as JSON")
  .action(async (opts: { weekOf?: string; json?: boolean }) => {
    try {
      const config = await loadConfig();
      const http = createHttpClient(config.access_token);
      const summary = await runWeekly(http, config.account_id, config.business_id, {
        weekOf: opts.weekOf,
      });

      if (opts.json) {
        console.log(formatJson(summary));
      } else {
        console.log(formatTable(summary));
      }
    } catch (err) {
      if (err instanceof AuthError) {
        console.error("Error: Token expired. Run `freshtime setup` to reconfigure.");
      } else {
        console.error("Error:", (err as Error).message);
      }
      process.exit(1);
    }
  });

program
  .command("clients")
  .description("List clients with their IDs")
  .action(async () => {
    try {
      const config = await loadConfig();
      const http = createHttpClient(config.access_token);
      console.log(await runClients(http, config.account_id));
    } catch (err) {
      if (err instanceof AuthError) {
        console.error("Error: Token expired. Run `freshtime setup` to reconfigure.");
      } else {
        console.error("Error:", (err as Error).message);
      }
      process.exit(1);
    }
  });

program
  .command("invoice <client-id>")
  .description("Create an invoice for all unbilled time entries for a client")
  .option("--rate <amount>", "Override the hourly rate for this run")
  .option("--currency <code>", "Override currency code (default: config or USD)")
  .option("--dry-run", "Show what would be invoiced without creating it")
  .option("--notes <text>", "Add notes to the invoice")
  .action(async (clientId: string, opts: { rate?: string; currency?: string; dryRun?: boolean; notes?: string }) => {
    try {
      const config = await loadConfig();
      const http = createHttpClient(config.access_token);
      const output = await runInvoice(http, config, parseInt(clientId, 10), {
        rate: opts.rate,
        currency: opts.currency,
        dryRun: opts.dryRun,
        notes: opts.notes,
      });
      console.log(output);
    } catch (err) {
      if (err instanceof AuthError) {
        console.error("Error: Token expired. Run `freshtime setup` to reconfigure.");
      } else {
        console.error("Error:", (err as Error).message);
      }
      process.exit(1);
    }
  });

program
  .command("refresh")
  .description("Refresh OAuth tokens (cron-friendly: silent on success, stderr on error)")
  .action(async () => {
    try {
      await runRefresh();
    } catch (err) {
      console.error("Error:", (err as Error).message);
      process.exit(1);
    }
  });

program.parse();
