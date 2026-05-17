import type { NextConfig } from "next";

const apiBaseUrl = (process.env.CODEMAP_API_BASE_URL ?? "http://localhost:18080").replace(/\/$/, "");

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${apiBaseUrl}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
