import { describe, test, expect, mock } from "bun:test";
import { runClients } from "../../src/commands/clients.ts";
import type { HttpClient } from "../../src/api/http.ts";

function mockHttp(clients: { id: number; organization: string; fname: string; lname: string }[]): HttpClient {
  return {
    get: mock(() => Promise.resolve({})) as HttpClient["get"],
    post: mock(() => Promise.resolve({})) as HttpClient["post"],
    put: mock(() => Promise.resolve({})) as HttpClient["put"],
    getPaginated: mock(() => Promise.resolve(clients)) as HttpClient["getPaginated"],
  };
}

describe("runClients", () => {
  test("formats clients as a table", async () => {
    const http = mockHttp([
      { id: 123, organization: "Acme Corp", fname: "", lname: "" },
      { id: 456, organization: "Widget Inc", fname: "", lname: "" },
    ]);

    const output = await runClients(http, "abc123");

    expect(output).toContain("ID");
    expect(output).toContain("Name");
    expect(output).toContain("123");
    expect(output).toContain("Acme Corp");
    expect(output).toContain("456");
    expect(output).toContain("Widget Inc");
  });

  test("shows message when no clients found", async () => {
    const http = mockHttp([]);
    const output = await runClients(http, "abc123");
    expect(output).toContain("No clients found");
  });

  test("uses fname/lname when organization is empty", async () => {
    const http = mockHttp([
      { id: 789, organization: "", fname: "John", lname: "Doe" },
    ]);

    const output = await runClients(http, "abc123");
    expect(output).toContain("789");
    expect(output).toContain("John Doe");
  });
});
