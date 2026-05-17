import type { NextConfig } from "next";
import { PHASE_DEVELOPMENT_SERVER } from "next/constants";

const apiBaseUrl = (process.env.CODEMAP_API_BASE_URL ?? "http://localhost:18080").replace(/\/$/, "");

const nextConfig = (phase: string): NextConfig => {
  const config: NextConfig = {
    output: "export",
  };

  if (phase === PHASE_DEVELOPMENT_SERVER) {
    config.rewrites = async () => [
      {
        source: "/api/:path*",
        destination: `${apiBaseUrl}/api/:path*`,
      },
    ];
  }

  return config;
};

export default nextConfig;
