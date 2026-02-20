import { describe, test, expect, mock } from "bun:test";
import { buildInvoiceLines, runInvoice } from "../../src/commands/invoice.ts";
import type { TimeEntry } from "../../src/api/time-entries.ts";
import type { HttpClient } from "../../src/api/http.ts";
import type { Config } from "../../src/config.ts";

const sampleEntries: TimeEntry[] = [
  {
    id: 1,
    client_id: 100,
    duration: 7200, // 2 hours
    started_at: "2026-02-09T09:00:00Z",
    local_started_at: "2026-02-09T09:00:00",
    note: "Frontend work",
    billable: true,
  },
  {
    id: 2,
    client_id: 100,
    duration: 5400, // 1.5 hours
    started_at: "2026-02-10T10:00:00Z",
    local_started_at: "2026-02-10T10:00:00",
    note: "",
    billable: true,
  },
];

describe("buildInvoiceLines", () => {
  test("creates one line per entry with correct fields", () => {
    const lines = buildInvoiceLines(sampleEntries, "150.00", "USD");

    expect(lines).toHaveLength(2);
    expect(lines[0]).toEqual({
      type: 0,
      name: "Frontend work",
      description: "2026-02-09",
      qty: "2.00",
      unit_cost: { amount: "150.00", code: "USD" },
    });
  });

  test("uses 'Consulting' when note is empty", () => {
    const lines = buildInvoiceLines(sampleEntries, "150.00", "USD");
    expect(lines[1]!.name).toBe("Consulting");
  });

  test("converts duration to hours with 2 decimal places", () => {
    const entries: TimeEntry[] = [
      {
        id: 1,
        client_id: 100,
        duration: 2700, // 0.75 hours
        started_at: "2026-02-09T09:00:00Z",
        local_started_at: "2026-02-09T09:00:00",
        note: "Task",
        billable: true,
      },
    ];
    const lines = buildInvoiceLines(entries, "100.00", "CAD");
    expect(lines[0]!.qty).toBe("0.75");
    expect(lines[0]!.unit_cost.code).toBe("CAD");
  });
});

describe("runInvoice", () => {
  const baseConfig: Config = {
    access_token: "test-token",
    account_id: "acc123",
    business_id: 456,
    client_rates: { "100": "150.00" },
    default_currency: "USD",
  };

  function mockHttp(entries: TimeEntry[]): HttpClient {
    return {
      get: mock(() => Promise.resolve({})) as HttpClient["get"],
      post: mock(() =>
        Promise.resolve({
          response: {
            result: {
              invoice: {
                invoiceid: 1,
                invoice_number: "0001",
                amount: { amount: "525.00", code: "USD" },
                links: { client_view: "https://my.freshbooks.com/view/abc" },
                v3_status: "sent",
              },
            },
          },
        })
      ) as HttpClient["post"],
      put: mock(() => Promise.resolve({})) as HttpClient["put"],
      getPaginated: mock(() => Promise.resolve(entries)) as HttpClient["getPaginated"],
    };
  }

  test("dry run shows summary without creating invoice", async () => {
    const http = mockHttp(sampleEntries);
    const output = await runInvoice(http, baseConfig, 100, { dryRun: true });

    expect(output).toContain("Dry run");
    expect(output).toContain("Entries:  2");
    expect(output).toContain("Hours:   3.50");
    expect(output).toContain("Rate:    150.00 USD/hr");
    expect(output).toContain("Total:   525.00 USD");
    expect(http.post).not.toHaveBeenCalled();
    expect(http.put).not.toHaveBeenCalled();
  });

  test("returns message when no unbilled entries", async () => {
    const http = mockHttp([]);
    const output = await runInvoice(http, baseConfig, 100, {});
    expect(output).toContain("No unbilled time entries");
  });

  test("throws when no rate configured", async () => {
    const http = mockHttp(sampleEntries);
    const configNoRate: Config = { ...baseConfig, client_rates: {} };

    expect(runInvoice(http, configNoRate, 100, {})).rejects.toThrow(
      "No rate configured"
    );
  });

  test("--rate flag overrides config rate", async () => {
    const http = mockHttp(sampleEntries);
    const output = await runInvoice(http, baseConfig, 100, {
      rate: "200.00",
      dryRun: true,
    });

    expect(output).toContain("Rate:    200.00 USD/hr");
    expect(output).toContain("Total:   700.00 USD");
  });

  test("creates invoice and marks entries billed", async () => {
    const http = mockHttp(sampleEntries);
    const output = await runInvoice(http, baseConfig, 100, {});

    expect(output).toContain("Invoice #0001 created");
    expect(output).toContain("https://my.freshbooks.com/view/abc");
    expect(http.post).toHaveBeenCalledTimes(1);
    expect(http.put).toHaveBeenCalledTimes(2); // 2 entries marked billed
  });
});
