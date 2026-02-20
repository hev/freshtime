import type { HttpClient } from "./http.ts";

interface ClientRecord {
  id: number;
  organization: string;
  fname: string;
  lname: string;
}

export async function listClients(
  http: HttpClient,
  accountId: string
): Promise<Map<number, string>> {
  const clients = await http.getPaginated<ClientRecord>(
    `/accounting/account/${accountId}/users/clients`,
    "clients"
  );

  const map = new Map<number, string>();
  for (const c of clients) {
    const name = c.organization || `${c.fname} ${c.lname}`.trim() || `Client #${c.id}`;
    map.set(c.id, name);
  }
  return map;
}
