import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Login, IDENTIFIER_REQUIRED, PASSWORD_REQUIRED } from "./Login";
import type { LoginRequest, LoginResult } from "./auth/api";

function setup(loginImpl?: (req: LoginRequest) => Promise<LoginResult>) {
  const login = vi.fn(loginImpl ?? (async () => ({ ok: true }) as LoginResult));
  const onSuccess = vi.fn();
  const user = userEvent.setup();
  render(<Login login={login} onSuccess={onSuccess} />);
  const identifier = screen.getByLabelText("Email or username");
  const password = screen.getByLabelText("Password");
  const submit = screen.getByRole("button", { name: /log in/i });
  return { login, onSuccess, user, identifier, password, submit };
}

describe("Login", () => {
  it("AC1: logs in with an email", async () => {
    const { login, onSuccess, user, identifier, password, submit } = setup();
    await user.type(identifier, "  jane@acme.shop ");
    await user.type(password, "s3cret");
    await user.click(submit);
    expect(login).toHaveBeenCalledWith({
      identifier: "jane@acme.shop",
      password: "s3cret",
      rememberMe: false,
    });
    expect(onSuccess).toHaveBeenCalledTimes(1);
  });

  it("AC2: logs in with a username and passes rememberMe", async () => {
    const { login, onSuccess, user, identifier, password, submit } = setup();
    await user.type(identifier, "jane");
    await user.type(password, "s3cret");
    await user.click(screen.getByLabelText("Remember me"));
    await user.click(submit);
    expect(login).toHaveBeenCalledWith({ identifier: "jane", password: "s3cret", rememberMe: true });
    expect(onSuccess).toHaveBeenCalledTimes(1);
  });

  it("AC3: shows an error for invalid credentials and does not redirect", async () => {
    const { onSuccess, user, identifier, password, submit } = setup(async () => ({
      ok: false,
      kind: "invalid-credentials",
      message: "Incorrect email/username or password.",
    }));
    await user.type(identifier, "jane");
    await user.type(password, "wrong");
    await user.click(submit);
    expect(await screen.findByRole("alert")).toHaveTextContent("Incorrect email/username or password.");
    expect(onSuccess).not.toHaveBeenCalled();
    expect(submit).toBeEnabled();
  });

  it("AC3: shows server-side field validation errors", async () => {
    const { user, identifier, password, submit } = setup(async () => ({
      ok: false,
      kind: "validation",
      message: "Please check the highlighted fields.",
      fieldErrors: { identifier: "Unknown format." },
    }));
    await user.type(identifier, "x");
    await user.type(password, "y");
    await user.click(submit);
    expect(await screen.findByText("Unknown format.")).toBeInTheDocument();
    expect(identifier).toHaveAttribute("aria-invalid", "true");
  });

  it("AC4: requires both fields and does not call the API", async () => {
    const { login, user, identifier, submit } = setup();
    await user.click(submit);
    expect(screen.getByText(IDENTIFIER_REQUIRED)).toBeInTheDocument();
    expect(screen.getByText(PASSWORD_REQUIRED)).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();

    await user.type(identifier, "   ");
    await user.click(submit);
    expect(screen.getByText(IDENTIFIER_REQUIRED)).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();
  });

  it("AC5: toggles password visibility", async () => {
    const { user, password } = setup();
    const toggle = screen.getByRole("button", { name: "Show password" });
    expect(password).toHaveAttribute("type", "password");
    expect(toggle).toHaveAttribute("aria-pressed", "false");
    await user.click(toggle);
    expect(password).toHaveAttribute("type", "text");
    expect(screen.getByRole("button", { name: "Hide password" })).toHaveAttribute("aria-pressed", "true");
    await user.click(screen.getByRole("button", { name: "Hide password" }));
    expect(password).toHaveAttribute("type", "password");
  });

  it("AC6: redirects to /dashboard by default on success", async () => {
    const assign = vi.fn();
    const original = window.location;
    Object.defineProperty(window, "location", { configurable: true, value: { ...original, assign } });
    try {
      const user = userEvent.setup();
      render(<Login login={async () => ({ ok: true })} />);
      await user.type(screen.getByLabelText("Email or username"), "jane");
      await user.type(screen.getByLabelText("Password"), "s3cret");
      await user.click(screen.getByRole("button", { name: /log in/i }));
      expect(assign).toHaveBeenCalledWith("/dashboard");
    } finally {
      Object.defineProperty(window, "location", { configurable: true, value: original });
    }
  });

  it("AC7: disables Login while the request is pending and blocks duplicates", async () => {
    let resolve!: (r: LoginResult) => void;
    const { login, onSuccess, user, identifier, password, submit } = setup(
      () => new Promise<LoginResult>((r) => (resolve = r)),
    );
    await user.type(identifier, "jane");
    await user.type(password, "s3cret");
    await user.click(submit);
    expect(submit).toBeDisabled();
    expect(submit).toHaveTextContent("Logging in…");
    await user.click(submit);
    await user.keyboard("{Enter}");
    expect(login).toHaveBeenCalledTimes(1);

    resolve({ ok: true });
    await vi.waitFor(() => expect(onSuccess).toHaveBeenCalledTimes(1));
  });

  it("AC7: keeps Login disabled after success while redirecting", async () => {
    const { login, user, identifier, password, submit } = setup();
    await user.type(identifier, "jane");
    await user.type(password, "s3cret");
    await user.click(submit);
    expect(submit).toBeDisabled();
    await user.keyboard("{Enter}");
    expect(login).toHaveBeenCalledTimes(1);
  });

  it("shows a generic error if the login call throws", async () => {
    const { onSuccess, user, identifier, password, submit } = setup(async () => {
      throw new Error("boom");
    });
    await user.type(identifier, "jane");
    await user.type(password, "s3cret");
    await user.click(submit);
    expect(await screen.findByRole("alert")).toHaveTextContent("Something went wrong");
    expect(onSuccess).not.toHaveBeenCalled();
    expect(submit).toBeEnabled();
  });

  it("offers a Forgot Password link", () => {
    setup();
    expect(screen.getByRole("link", { name: "Forgot password?" })).toHaveAttribute("href", "/forgot-password");
  });
});
