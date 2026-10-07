import Link from "next/link";

export default function LoginPage() {
  return (
    <main className="login-signup-page">
      <section className="auth-card" aria-labelledby="login-heading">
        <Link className="auth-brand" href="/store">GemBid</Link>
        <h1 id="login-heading">Log in</h1>
        <p className="auth-description">Welcome back. Log in to continue.</p>
        <form id="loginForm">
          <label htmlFor="email">Email</label>
          <input
            type="email"
            id="email"
            name="email"
            autoComplete="email"
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
          <button type="submit">Log in</button>
        </form>
        <p className="auth-switch">
          New to GemBid? <Link href="/signup">Create an account</Link>
        </p>
        <Link className="auth-store-link" href="/store">Return to store</Link>
      </section>
    </main>
  );
}