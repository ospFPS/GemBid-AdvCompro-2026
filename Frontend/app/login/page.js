import Link from "next/link";

export default function LoginPage() {
  return (
    <main className="login-signup-page">
      <form id="loginForm">
        <h1>Log In</h1>
        {/* Input */}
        <label htmlFor="username">Username</label>
        <input type="text" id="username" name="username" required />
        <label htmlFor="password">Password</label>
        <input type="password" id="password" name="password" required />
        <button type="submit">Log In</button>
        {/* Link */}
        <Link className="nav-link" href="/signup">Sign up</Link>
      </form>
    </main>
  );
}