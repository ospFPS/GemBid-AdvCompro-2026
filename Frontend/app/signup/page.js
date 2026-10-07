import Link from "next/link";

export default function SignupPage() {
  return (
    <main className="login-signup-page">
      <form id="signupForm">
        <h1>Sign Up</h1>
        {/* Input */}
        <label htmlFor="email">Email</label>
        <input type="text" id="email" name="email" required />
        <label htmlFor="password">Password</label>
        <input type="password" id="password" name="password" required />
        <label htmlFor="passwordConf">Password confirmation</label>
        <input type="password" id="passwordConf" name="passwordConf" required />
        <button type="submit">Sign up</button>
        {/* Link */}
        <Link className="nav-link" href="/login">Return to log in</Link>
      </form>
    </main>
  );
}