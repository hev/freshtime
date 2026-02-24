export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly statusText: string,
    public readonly body: string
  ) {
    super(`API error ${status} ${statusText}: ${body}`);
    this.name = "ApiError";
  }
}

export class AuthError extends ApiError {
  constructor(body: string) {
    super(401, "Unauthorized", body);
    this.name = "AuthError";
  }
}

export interface HttpClient {
  get<T>(path: string, params?: Record<string, string>): Promise<T>;
  post<T>(path: string, body: unknown): Promise<T>;
  put<T>(path: string, body: unknown): Promise<T>;
  getPaginated<T>(
    path: string,
    resultKey: string,
    params?: Record<string, string>
  ): Promise<T[]>;
}

import { loadConfig, saveConfig } from "../config.ts";

const BASE_URL = "https://api.freshbooks.com";

export async function refreshAccessToken(refreshToken: string): Promise<{ access_token: string; refresh_token: string }> {
  const res = await fetch("https://api.freshbooks.com/auth/oauth/token", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      grant_type: "refresh_token",
      client_id: process.env.FRESHBOOKS_CLIENT_ID,
      client_secret: process.env.FRESHBOOKS_CLIENT_SECRET,
      refresh_token: refreshToken,
    }),
  });

  if (!res.ok) {
    throw new Error(`Token refresh failed (${res.status})`);
  }

  const data = (await res.json()) as { access_token: string; refresh_token: string };
  return { access_token: data.access_token, refresh_token: data.refresh_token };
}

export function createHttpClient(accessToken: string): HttpClient {
  let currentToken = accessToken;
  let hasRetried = false;

  async function handleRefresh(): Promise<void> {
    const config = await loadConfig();
    if (!config.refresh_token) {
      throw new AuthError("No refresh token available. Run `freshtime setup` to re-authenticate.");
    }

    const tokens = await refreshAccessToken(config.refresh_token);
    currentToken = tokens.access_token;
    await saveConfig({
      ...config,
      access_token: tokens.access_token,
      refresh_token: tokens.refresh_token,
    });
  }

  async function request<T>(
    path: string,
    params?: Record<string, string>
  ): Promise<T> {
    const url = new URL(path, BASE_URL);
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        url.searchParams.set(key, value);
      }
    }

    const res = await fetch(url.toString(), {
      headers: {
        Authorization: `Bearer ${currentToken}`,
        "Content-Type": "application/json",
      },
    });

    const body = await res.text();

    if (res.status === 401 && !hasRetried) {
      hasRetried = true;
      try {
        await handleRefresh();
        return request<T>(path, params);
      } catch {
        throw new AuthError("Session expired. Run `freshtime setup` to re-authenticate.");
      }
    }

    if (res.status === 401) {
      throw new AuthError(body);
    }
    if (!res.ok) {
      throw new ApiError(res.status, res.statusText, body);
    }

    return JSON.parse(body) as T;
  }

  async function mutate<T>(
    method: "POST" | "PUT",
    path: string,
    body: unknown
  ): Promise<T> {
    const url = new URL(path, BASE_URL);

    const res = await fetch(url.toString(), {
      method,
      headers: {
        Authorization: `Bearer ${currentToken}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    const responseBody = await res.text();

    if (res.status === 401 && !hasRetried) {
      hasRetried = true;
      try {
        await handleRefresh();
        return mutate<T>(method, path, body);
      } catch {
        throw new AuthError("Session expired. Run `freshtime setup` to re-authenticate.");
      }
    }

    if (res.status === 401) {
      throw new AuthError(responseBody);
    }
    if (!res.ok) {
      throw new ApiError(res.status, res.statusText, responseBody);
    }

    return JSON.parse(responseBody) as T;
  }

  return {
    get: request,
    post: <T>(path: string, body: unknown) => mutate<T>("POST", path, body),
    put: <T>(path: string, body: unknown) => mutate<T>("PUT", path, body),

    async getPaginated<T>(
      path: string,
      resultKey: string,
      params?: Record<string, string>
    ): Promise<T[]> {
      const allResults: T[] = [];
      let page = 1;
      let totalPages = 1;

      do {
        const pageParams = { ...params, page: String(page), per_page: "100" };
        const data = await request<Record<string, unknown>>(path, pageParams);

        // FreshBooks APIs use two different response shapes:
        // - Accounting: { response: { result: { [key]: [...], pages, ... } } }
        // - Timetracking: { [key]: [...], meta: { pages, ... } }
        let items: T[];
        if (data[resultKey]) {
          // Top-level shape (timetracking)
          items = data[resultKey] as T[];
          const meta = data.meta as Record<string, unknown> | undefined;
          totalPages = (meta?.pages as number) ?? 1;
        } else {
          // Nested shape (accounting)
          const response = data.response as Record<string, unknown> | undefined;
          const result = response?.result as Record<string, unknown> | undefined;
          items = (result?.[resultKey] ?? []) as T[];
          totalPages = (result?.pages as number) ?? 1;
        }

        allResults.push(...items);
        page++;
      } while (page <= totalPages);

      return allResults;
    },
  };
}
