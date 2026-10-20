"use client";

import Link from "next/link";
import { useState } from "react";
import { api, saveLogin } from "../lib/api";

export default function SignupPage() {
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function handleSubmit(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    if (form.get("password") !== form.get("passwordConf")) {
      setError("Passwords do not match.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const credentials = { username: form.get("username"), password: form.get("password") };
      await api("/api/register", { method: "POST", body: { ...credentials, email: form.get("email") } });
      const { token, user } = await api("/api/login", { method: "POST", body: credentials });
      saveLogin(token, user);
      window.location.href = "/";
    } catch (err) {
      setError(err.message);
      setBusy(false);
    }
  }

  return (
    <main className="login-signup-page">
      <section className="auth-card" aria-labelledby="signup-heading">
        <Link className="auth-brand" href="/store">GemBid</Link>
        <h1 id="signup-heading">Create your account</h1>
        <p className="auth-description">Sign up to start exploring the store.</p>
        <form id="signupForm" onSubmit={handleSubmit}>
          <label htmlFor="username">Username</label>
          <input
            type="text"
            id="username"
            name="username"
            autoComplete="username"
            pattern="[a-z0-9._\-]{3,30}"
            title="3-30 characters: lowercase letters, numbers, dot, underscore or hyphen"
            required
          />
          <label htmlFor="email">Email (optional)</label>
          <input
            type="email"
            id="email"
            name="email"
            autoComplete="email"
          />
          <label htmlFor="password">Password</label>
          <input
            type="password"
            id="password"
            name="password"
            autoComplete="new-password"
            minLength={8}
            maxLength={72}
            required
          />
          <label htmlFor="passwordConf">Confirm password</label>
          <input
            type="password"
            id="passwordConf"
            name="passwordConf"
            autoComplete="new-password"
            required
          />
          {error && <p className="auth-error" role="alert">{error}</p>}
          <button type="submit" disabled={busy}>{busy ? "Creating account..." : "Create account"}</button>
        </form>
        <p className="auth-switch">
          Already have an account? <Link href="/login">Log in</Link>
        </p>
        <Link className="auth-store-link" href="/store">Return to store</Link>
      </section>
    </main>
  );
}