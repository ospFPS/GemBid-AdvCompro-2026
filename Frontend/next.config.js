const path = require("node:path");

// Backend the /api proxy points at. Set at build time (rewrites are baked into the build).
const backend = process.env.BACKEND_URL ?? "http://localhost:8080";

/** @type {import('next').NextConfig} */
const nextConfig = {
  outputFileTracingRoot: path.join(__dirname),
  // The Go backend sends no CORS headers, so the browser only talks to this origin.
  rewrites() {
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
};

module.exports = nextConfig;
