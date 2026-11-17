import { afterEach, describe, expect, it, vi } from "vitest";
import { GENERIC_ERROR_MESSAGE, INVALID_CREDENTIALS_MESSAGE, login } from "./api";

const req = { identifier: "jane@acme.shop", password: "s3cret", rememberMe: true };

function mockFetch(impl: () => Promise<Response>) {
  const fn = vi.fn(impl);
  vi.stubGlobal("fetch", fn);
  return fn;
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

describe("auth api login", () => {
  it("posts credentials with cookies included", async () => {
    vi.stubEnv("VITE_AUTH_API_URL", "https://auth.example.com/");
    const fetchFn = mockFetch(async () => new Response(null, { status: 200 }));
    await expect(login(req)).resolves.toEqual({ ok: true });
    expect(fetchFn).toHaveBeenCalledWith("https://auth.example.com/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(req),
    });
  });

  it("maps 401 to invalid credentials", async () => {
    mockFetch(async () => new Response(null, { status: 401 }));
    await expect(login(req)).resolves.toEqual({
      ok: false,
      kind: "invalid-credentials",
      message: INVALID_CREDENTIALS_MESSAGE,
    });
  });

  it("maps 400 to field errors", async () => {
    mockFetch(
      async () =>
        new Response(JSON.stringify({ message: "Bad input", errors: { password: "Too short", other: "x" } }), {
          status: 400,
        }),
    );
    await expect(login(req)).resolves.toEqual({
      ok: false,
      kind: "validation",
      message: "Bad input",
      fieldErrors: { password: "Too short" },
    });
  });

  it("handles 400 with a non-JSON body", async () => {
    mockFetch(async () => new Response("nope", { status: 400 }));
    const result = await login(req);
    expect(result).toMatchObject({ ok: false, kind: "validation", fieldErrors: {} });
  });

  it("maps 5xx and network failures to a generic error", async () => {
    mockFetch(async () => new Response(null, { status: 503 }));
    await expect(login(req)).resolves.toEqual({ ok: false, kind: "error", message: GENERIC_ERROR_MESSAGE });
    mockFetch(async () => {
      throw new TypeError("network");
    });
    await expect(login(req)).resolves.toEqual({ ok: false, kind: "error", message: GENERIC_ERROR_MESSAGE });
  });
});
