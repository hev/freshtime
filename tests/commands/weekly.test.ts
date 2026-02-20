import { describe, test, expect } from "bun:test";
import { getWeekRange, buildSummary } from "../../src/commands/weekly.ts";
import type { TimeEntry } from "../../src/api/time-entries.ts";

describe("getWeekRange", () => {
  test("returns Monday–Friday for a Wednesday", () => {
    const result = getWeekRange(new Date("2026-02-11T12:00:00"));
    expect(result.weekStart).toBe("2026-02-09");
    expect(result.weekEnd).toBe("2026-02-13");
  });

  test("returns same week for a Monday", () => {
    const result = getWeekRange(new Date("2026-02-09T00:00:00"));
    expect(result.weekStart).toBe("2026-02-09");
    expect(result.weekEnd).toBe("2026-02-13");
  });

  test("returns same week for a Friday", () => {
    const result = getWeekRange(new Date("2026-02-13T23:59:59"));
    expect(result.weekStart).toBe("2026-02-09");
    expect(result.weekEnd).toBe("2026-02-13");
  });

  test("handles Sunday (goes to previous week's Monday)", () => {
    const result = getWeekRange(new Date("2026-02-15T12:00:00"));
    expect(result.weekStart).toBe("2026-02-09");
    expect(result.weekEnd).toBe("2026-02-13");
  });

  test("handles Saturday", () => {
    const result = getWeekRange(new Date("2026-02-14T12:00:00"));
    expect(result.weekStart).toBe("2026-02-09");
    expect(result.weekEnd).toBe("2026-02-13");
  });

  test("handles month boundary", () => {
    const result = getWeekRange(new Date("2026-03-02T12:00:00"));
    expect(result.weekStart).toBe("2026-03-02");
    expect(result.weekEnd).toBe("2026-03-06");
  });

  test("handles year boundary", () => {
    const result = getWeekRange(new Date("2025-12-31T12:00:00"));
    expect(result.weekStart).toBe("2025-12-29");
    expect(result.weekEnd).toBe("2026-01-02");
  });
});

describe("buildSummary", () => {
  const clientNames = new Map<number, string>([
    [1, "Acme Corp"],
    [2, "Globex Inc"],
  ]);

  test("groups entries by client correctly", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 1, duration: 7200, started_at: "2026-02-09T09:00:00Z" },
      { id: 2, client_id: 1, duration: 3600, started_at: "2026-02-10T10:00:00Z" },
      { id: 3, client_id: 2, duration: 5400, started_at: "2026-02-09T14:00:00Z" },
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");

    expect(summary.clients).toHaveLength(2);
    const acme = summary.clients.find((c) => c.name === "Acme Corp")!;
    expect(acme.daily[0]).toBe(2); // Mon: 7200s = 2h
    expect(acme.daily[1]).toBe(1); // Tue: 3600s = 1h
    expect(acme.total).toBe(3);

    const globex = summary.clients.find((c) => c.name === "Globex Inc")!;
    expect(globex.daily[0]).toBe(1.5); // Mon: 5400s = 1.5h
    expect(globex.total).toBe(1.5);
  });

  test("handles zero-entry weeks", () => {
    const summary = buildSummary([], clientNames, "2026-02-09");
    expect(summary.clients).toHaveLength(0);
    expect(summary.grandTotal).toBe(0);
    expect(summary.weekStart).toBe("2026-02-09");
    expect(summary.weekEnd).toBe("2026-02-13");
  });

  test("converts duration seconds to hours correctly", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 1, duration: 5400, started_at: "2026-02-09T09:00:00Z" }, // 1.5h
      { id: 2, client_id: 1, duration: 900, started_at: "2026-02-10T10:00:00Z" },  // 0.25h
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");
    const acme = summary.clients[0]!;
    expect(acme.daily[0]).toBe(1.5);
    expect(acme.daily[1]).toBe(0.25);
    expect(acme.total).toBe(1.75);
  });

  test("handles unknown client_id", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 999, duration: 3600, started_at: "2026-02-09T09:00:00Z" },
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");
    expect(summary.clients[0]!.name).toBe("Client #999");
  });

  test("sums multiple entries for same client on same day", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 1, duration: 3600, started_at: "2026-02-09T09:00:00Z" },
      { id: 2, client_id: 1, duration: 3600, started_at: "2026-02-09T14:00:00Z" },
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");
    const acme = summary.clients[0]!;
    expect(acme.daily[0]).toBe(2); // 2 x 1h = 2h
    expect(acme.total).toBe(2);
  });

  test("skips weekend entries", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 1, duration: 3600, started_at: "2026-02-14T09:00:00Z" }, // Saturday
      { id: 2, client_id: 1, duration: 3600, started_at: "2026-02-15T09:00:00Z" }, // Sunday
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");
    expect(summary.clients).toHaveLength(0);
    expect(summary.grandTotal).toBe(0);
  });

  test("sorts clients alphabetically", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 2, duration: 3600, started_at: "2026-02-09T09:00:00Z" },
      { id: 2, client_id: 1, duration: 3600, started_at: "2026-02-09T10:00:00Z" },
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");
    expect(summary.clients[0]!.name).toBe("Acme Corp");
    expect(summary.clients[1]!.name).toBe("Globex Inc");
  });

  test("calculates grandTotal correctly across clients", () => {
    const entries: TimeEntry[] = [
      { id: 1, client_id: 1, duration: 7200, started_at: "2026-02-09T09:00:00Z" },
      { id: 2, client_id: 2, duration: 5400, started_at: "2026-02-10T10:00:00Z" },
    ];

    const summary = buildSummary(entries, clientNames, "2026-02-09");
    expect(summary.grandTotal).toBe(3.5); // 2h + 1.5h
  });
});
