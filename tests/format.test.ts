import { describe, test, expect } from "bun:test";
import { formatTable, formatJson, type WeeklySummary } from "../src/format.ts";

const sampleSummary: WeeklySummary = {
  weekStart: "2026-02-09",
  weekEnd: "2026-02-13",
  clients: [
    { name: "Acme Corp", daily: [2.0, 3.5, 4.0, 2.0, 1.5], total: 13.0 },
    { name: "Globex Inc", daily: [1.0, 0, 2.0, 3.0, 2.0], total: 8.0 },
  ],
  grandTotal: 21.0,
};

describe("formatTable", () => {
  test("contains week header", () => {
    const output = formatTable(sampleSummary);
    expect(output).toContain("Week of Feb 9 – Feb 13, 2026");
  });

  test("contains column headers", () => {
    const output = formatTable(sampleSummary);
    expect(output).toContain("Client");
    expect(output).toContain("Mon");
    expect(output).toContain("Fri");
    expect(output).toContain("Total");
  });

  test("contains client rows with hours", () => {
    const output = formatTable(sampleSummary);
    expect(output).toContain("Acme Corp");
    expect(output).toContain("13.0h");
    expect(output).toContain("Globex Inc");
    expect(output).toContain("8.0h");
  });

  test("shows dash for zero hours", () => {
    const output = formatTable(sampleSummary);
    // Globex has 0 on Tuesday
    expect(output).toContain("—");
  });

  test("contains grand total", () => {
    const output = formatTable(sampleSummary);
    const lines = output.split("\n");
    const totalLine = lines.find((l) => l.startsWith("Total"));
    expect(totalLine).toBeDefined();
    expect(totalLine).toContain("21.0h");
  });

  test("renders empty week correctly", () => {
    const empty: WeeklySummary = {
      weekStart: "2026-02-09",
      weekEnd: "2026-02-13",
      clients: [],
      grandTotal: 0,
    };
    const output = formatTable(empty);
    expect(output).toContain("Week of Feb 9 – Feb 13, 2026");
    expect(output).toContain("Total");
    expect(output).toContain("—h");
  });
});

describe("formatJson", () => {
  test("outputs valid JSON", () => {
    const output = formatJson(sampleSummary);
    const parsed = JSON.parse(output);
    expect(parsed).toBeDefined();
  });

  test("preserves structure", () => {
    const output = formatJson(sampleSummary);
    const parsed = JSON.parse(output);
    expect(parsed.weekStart).toBe("2026-02-09");
    expect(parsed.weekEnd).toBe("2026-02-13");
    expect(parsed.clients).toHaveLength(2);
    expect(parsed.grandTotal).toBe(21.0);
  });

  test("includes client details", () => {
    const output = formatJson(sampleSummary);
    const parsed = JSON.parse(output);
    expect(parsed.clients[0].name).toBe("Acme Corp");
    expect(parsed.clients[0].daily).toEqual([2.0, 3.5, 4.0, 2.0, 1.5]);
    expect(parsed.clients[0].total).toBe(13.0);
  });

  test("handles empty week", () => {
    const empty: WeeklySummary = {
      weekStart: "2026-02-09",
      weekEnd: "2026-02-13",
      clients: [],
      grandTotal: 0,
    };
    const output = formatJson(empty);
    const parsed = JSON.parse(output);
    expect(parsed.clients).toEqual([]);
    expect(parsed.grandTotal).toBe(0);
  });
});
