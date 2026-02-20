import { createHttpClient } from "../api/http.ts";
import { getIdentity } from "../api/identity.ts";
import { saveConfig, getConfigPath } from "../config.ts";

const CLIENT_ID = process.env.FRESHBOOKS_CLIENT_ID!;
const CLIENT_SECRET = process.env.FRESHBOOKS_CLIENT_SECRET!;
const REDIRECT_URI = "https://localhost:8457/callback";
const AUTH_URL = `https://auth.freshbooks.com/service/auth/oauth/authorize?client_id=${CLIENT_ID}&response_type=code&redirect_uri=${encodeURIComponent(REDIRECT_URI)}`;

async function generateSelfSignedCert() {
  const proc = Bun.spawn([
    "openssl", "req", "-x509", "-newkey", "rsa:2048",
    "-keyout", "/dev/stdout", "-out", "/dev/stdout",
    "-days", "1", "-nodes",
    "-subj", "/CN=localhost",
  ], { stdout: "pipe", stderr: "pipe" });

  const output = await new Response(proc.stdout).text();
  await proc.exited;

  const certMatch = output.match(/(-----BEGIN CERTIFICATE-----[\s\S]+?-----END CERTIFICATE-----)/);
  const keyMatch = output.match(/(-----BEGIN PRIVATE KEY-----[\s\S]+?-----END PRIVATE KEY-----)/);

  if (!certMatch || !keyMatch) {
    throw new Error("Failed to generate self-signed certificate.");
  }

  return { cert: certMatch[1], key: keyMatch[1] };
}

async function waitForAuthCode(): Promise<string> {
  const { cert, key } = await generateSelfSignedCert();

  return new Promise((resolve, reject) => {
    const server = Bun.serve({
      port: 8457,
      tls: { cert, key },
      routes: {
        "/callback": (req) => {
          const url = new URL(req.url);
          const code = url.searchParams.get("code");

          if (!code) {
            reject(new Error("No authorization code received."));
            return new Response("Error: no code received. Close this tab and try again.", {
              status: 400,
            });
          }

          resolve(code);
          setTimeout(() => server.stop(), 100);

          return new Response(
            "<html><body><h1>Done! You can close this tab.</h1></body></html>",
            { headers: { "Content-Type": "text/html" } }
          );
        },
      },
      fetch(req) {
        return new Response("Not found", { status: 404 });
      },
    });
  });
}

async function exchangeCodeForToken(code: string): Promise<string> {
  const res = await fetch("https://api.freshbooks.com/auth/oauth/token", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      grant_type: "authorization_code",
      client_id: CLIENT_ID,
      client_secret: CLIENT_SECRET,
      code,
      redirect_uri: REDIRECT_URI,
    }),
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(`Token exchange failed (${res.status}): ${body}`);
  }

  const data = (await res.json()) as { access_token: string };
  return data.access_token;
}

export async function runSetup(): Promise<void> {
  if (!CLIENT_ID || !CLIENT_SECRET) {
    console.error("Missing FRESHBOOKS_CLIENT_ID or FRESHBOOKS_CLIENT_SECRET in .env");
    process.exit(1);
  }

  console.log("Open this link to authorize freshtime:\n");
  console.log(`  ${AUTH_URL}\n`);

  console.log("Waiting for authorization...");
  const code = await waitForAuthCode();

  console.log("Exchanging code for token...");
  const accessToken = await exchangeCodeForToken(code);

  console.log("Verifying token...");
  const http = createHttpClient(accessToken);
  const identity = await getIdentity(http);

  await saveConfig({
    access_token: accessToken,
    account_id: identity.account_id,
    business_id: identity.business_id,
  });

  console.log(`\nSetup complete.`);
  console.log(`  Account: ${identity.account_id}`);
  console.log(`  Business: ${identity.business_id}`);
  console.log(`  Config saved to: ${getConfigPath()}`);
}
