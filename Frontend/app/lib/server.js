import { connection } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

// serverApi calls the backend from server components. Returns null when the backend errors.
export async function serverApi(path) {
	await connection(); // render per request, never at build time
	try {
		const res = await fetch(`${BACKEND_URL}${path}`, { cache: "no-store" });
		return res.ok ? res.json() : null;
	} catch {
		return null;
	}
}
