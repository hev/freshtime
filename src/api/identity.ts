import type { HttpClient } from "./http.ts";

interface MeResponse {
  response: {
    id: number;
    business_memberships: Array<{
      business: {
        id: number;
        account_id: string;
      };
    }>;
  };
}

export interface Identity {
  account_id: string;
  business_id: number;
}

export async function getIdentity(http: HttpClient): Promise<Identity> {
  const data = await http.get<MeResponse>("/auth/api/v1/users/me");
  const memberships = data.response.business_memberships;

  if (memberships.length === 0) {
    throw new Error("No business memberships found on this account.");
  }

  const first = memberships[0]!;
  return {
    account_id: first.business.account_id,
    business_id: first.business.id,
  };
}
