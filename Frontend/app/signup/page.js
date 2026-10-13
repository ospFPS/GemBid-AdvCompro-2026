import Link from "next/link";

export default function SignupPage() {
  return (
    <main className="login-signup-page">
      <section className="auth-card" aria-labelledby="signup-heading">
        <Link className="auth-brand" href="/store">GemBid</Link>
        <h1 id="signup-heading">Create your account</h1>
        <p className="auth-description">Sign up to start exploring the store.</p>
        <form id="signupForm">
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
            autoComplete="new-password"
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
          <button type="submit">Create account</button>
        </form>
        <p className="auth-switch">
          Already have an account? <Link href="/login">Log in</Link>
        </p>
        <Link className="auth-store-link" href="/store">Return to store</Link>
      </section>
    </main>
  );
}