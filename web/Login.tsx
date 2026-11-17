import { useRef, useState, type FormEvent } from "react";
import { login as defaultLogin, type FieldErrors, type LoginRequest, type LoginResult } from "./auth/api";

export const IDENTIFIER_REQUIRED = "Email or username is required.";
export const PASSWORD_REQUIRED = "Password is required.";

type LoginProps = {
  /** Called after successful authentication. Defaults to redirecting to /dashboard. */
  onSuccess?: () => void;
  /** Injected for tests; defaults to the auth service client. */
  login?: (req: LoginRequest) => Promise<LoginResult>;
};

function redirectToDashboard() {
  window.location.assign("/dashboard");
}

export function Login({ onSuccess = redirectToDashboard, login = defaultLogin }: LoginProps) {
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  // Ref guard blocks a second submit before React re-renders the disabled button.
  const inFlight = useRef(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (inFlight.current) return;

    const trimmedIdentifier = identifier.trim();
    const errors: FieldErrors = {};
    if (!trimmedIdentifier) errors.identifier = IDENTIFIER_REQUIRED;
    if (!password) errors.password = PASSWORD_REQUIRED;
    setFieldErrors(errors);
    setFormError("");
    if (errors.identifier || errors.password) return;

    inFlight.current = true;
    setSubmitting(true);
    try {
      const result = await login({ identifier: trimmedIdentifier, password, rememberMe });
      if (result.ok) {
        onSuccess();
        return;
      }
      if (result.kind === "validation") setFieldErrors(result.fieldErrors);
      setFormError(result.message);
    } finally {
      inFlight.current = false;
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} noValidate aria-label="Log in">
      <h1>Log in</h1>

      {formError && <p role="alert">{formError}</p>}

      <div>
        <label htmlFor="login-identifier">Email or username</label>
        <input
          id="login-identifier"
          name="identifier"
          type="text"
          autoComplete="username"
          value={identifier}
          onChange={(e) => setIdentifier(e.target.value)}
          aria-invalid={fieldErrors.identifier ? true : undefined}
          aria-describedby={fieldErrors.identifier ? "login-identifier-error" : undefined}
          disabled={submitting}
        />
        {fieldErrors.identifier && <p id="login-identifier-error">{fieldErrors.identifier}</p>}
      </div>

      <div>
        <label htmlFor="login-password">Password</label>
        <input
          id="login-password"
          name="password"
          type={showPassword ? "text" : "password"}
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          aria-invalid={fieldErrors.password ? true : undefined}
          aria-describedby={fieldErrors.password ? "login-password-error" : undefined}
          disabled={submitting}
        />
        <button
          type="button"
          onClick={() => setShowPassword((v) => !v)}
          aria-pressed={showPassword}
          aria-controls="login-password"
        >
          {showPassword ? "Hide password" : "Show password"}
        </button>
        {fieldErrors.password && <p id="login-password-error">{fieldErrors.password}</p>}
      </div>

      <div>
        <label>
          <input
            type="checkbox"
            name="rememberMe"
            checked={rememberMe}
            onChange={(e) => setRememberMe(e.target.checked)}
            disabled={submitting}
          />
          Remember me
        </label>
        <a href="/forgot-password">Forgot password?</a>
      </div>

      <button type="submit" disabled={submitting} aria-busy={submitting}>
        {submitting ? "Logging in…" : "Log in"}
      </button>
    </form>
  );
}
