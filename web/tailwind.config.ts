import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Aptos", "IBM Plex Sans", "Segoe UI", "sans-serif"],
        mono: ["JetBrains Mono", "IBM Plex Mono", "Menlo", "monospace"],
      },
      colors: {
        ink: "#17202a",
        paper: "#f7f5ef",
        steel: "#52616b",
        moss: "#58745f",
        signal: "#d95f3d",
        line: "#d9d3c7",
      },
    },
  },
  plugins: [],
};

export default config;
