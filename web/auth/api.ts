// Client for the existing authentication service.
//
// Assumed contract (confirm with the auth service owners):
//   POST {VITE_AUTH_API_URL}/login
//   body: { identifier, password, rememberMe }
//   200 -> authenticated; the service sets an HttpOnly session cookie
//   400 -> { errors?: { identifier?: string; password?: string }, message?: string }
//   401 -> invalid credentials
// The session token never reaches JavaScript; requests use credentials: "include".

export type LoginRequest = {
  identifier: string;
  password: string;
  rememberMe: boolean;
};

export type FieldErrors = Partial<Record<"identifier" | "password", string>>;

export type LoginResult =
  | { ok: true }
  | { ok: false; kind: "invalid-credentials"; message: string }
  | { ok: false; kind: "validation"; message: string; fieldErrors: FieldErrors }
  | { ok: false; kind: "error"; message: string };

export const INVALID_CREDENTIALS_MESSAGE =
  "Incorrect email/username or password.";
export const GENERIC_ERROR_MESSAGE =
  "Something went wrong. Please try again.";

function authBaseUrl(): string {
  return (import.meta.env.VITE_AUTH_API_URL ?? "").replace(/\/+$/, "");
}

export async function login(req: LoginRequest): Promise<LoginResult> {
  let res: Response;
  try {
    res = await fetch(`${authBaseUrl()}/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify(req),
    });
  } catch {
    return { ok: false, kind: "error", message: GENERIC_ERROR_MESSAGE };
  }

  if (res.ok) return { ok: true };

  if (res.status === 401) {
    return {
      ok: false,
      kind: "invalid-credentials",
      message: INVALID_CREDENTIALS_MESSAGE,
    };
  }

  if (res.status === 400) {
    const body = await res.json().catch(() => ({}));
    const errors = (body && typeof body === "object" ? body.errors : undefined) ?? {};
    const fieldErrors: FieldErrors = {};
    if (typeof errors.identifier === "string") fieldErrors.identifier = errors.identifier;
    if (typeof errors.password === "string") fieldErrors.password = errors.password;
    const message =
      typeof body?.message === "string" ? body.message : "Please check the highlighted fields.";
    return { ok: false, kind: "validation", message, fieldErrors };
  }

  return { ok: false, kind: "error", message: GENERIC_ERROR_MESSAGE };
}
