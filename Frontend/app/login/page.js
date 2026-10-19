"use client";

import Link from "next/link";
import { useState } from "react";
import { api, saveLogin } from "../lib/api";

export default function LoginPage() {
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function handleSubmit(event) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setBusy(true);
    setError("");
    try {
      const { token, user } = await api("/api/login", {
        method: "POST",
        body: { username: form.get("username"), password: form.get("password") },
      });
      saveLogin(token, user);
      window.location.href = "/store";
    } catch (err) {
      setError(err.message);
      setBusy(false);
    }
  }

  return (
    <main className="login-signup-page">
      <img src="/assets/Dependencies2.gif"/>
      <section className="auth-card" aria-labelledby="login-heading">
        <Link className="auth-brand" href="/store">GemBid</Link>
        <h1 id="login-heading">Log in</h1>
        <p className="auth-description">Welcome back. Log in to continue.</p>
        <form id="loginForm" onSubmit={handleSubmit}>
          <label htmlFor="username">Username</label>
          <input
            type="text"
            id="username"
            name="username"
            autoComplete="username"
            required
          />
          <label htmlFor="password">Password</label>
          <input
            type="password"
            id="password"
            name="password"
            autoComplete="current-password"
            required
          />
          {error && <p className="auth-error" role="alert">{error}</p>}
          <button type="submit" disabled={busy}>{busy ? "Logging in..." : "Log in"}</button>
        </form>
        <p className="auth-switch">
          New to GemBid? <Link href="/signup">Create an account</Link>
        </p>
        <Link className="auth-store-link" href="/store">Return to store</Link>
      </section>
    </main>
  );
}