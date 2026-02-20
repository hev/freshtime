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

const BASE_URL = "https://api.freshbooks.com";

export function createHttpClient(accessToken: string): HttpClient {
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
        Authorization: `Bearer ${accessToken}`,
        "Content-Type": "application/json",
      },
    });

    const body = await res.text();

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
        Authorization: `Bearer ${accessToken}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    const responseBody = await res.text();

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
