import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/web_ui/src/**/*.{js,ts,jsx,tsx,mdx}"
  ],
  theme: {
    extend: {
      colors: {
        background: "#0B0E14",
        surface: {
          DEFAULT: "#141923",
          light: "#1F2633",
          card: "#111620",
        },
        brand: {
          navy: "#0A192F",
          green: "#00F0A0",
          red: "#FF3B56",
          gold: "#F5A623",
        },
        border: "hsl(215 27.9% 16.9%)",
        input: "hsl(215 27.9% 16.9%)",
        ring: "hsl(160 100% 47%)",
      },
      fontFamily: {
        mono: ["JetBrains Mono", "monospace"],
        sans: ["Inter", "sans-serif"],
      },
    },
  },
  plugins: [],
};
export default config;
