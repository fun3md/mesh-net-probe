/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        "primary": "#135bec",
        "background-light": "#f6f6f8",
        "background-dark": "#101622",
        "panel-light": "#ffffff",
        "panel-dark": "#182134",
        "border-light": "#e5e7eb",
        "border-dark": "#3b4354",
        "text-light-primary": "#111827",
        "text-dark-primary": "#ffffff",
        "text-light-secondary": "#6b7280",
        "text-dark-secondary": "#9da6b9",
        "success": "#28A745",
        "warning": "#FFC107",
        "error": "#DC3545",
      },
      fontFamily: {
        "display": ["Inter", "sans-serif"]
      },
      borderRadius: {
        "DEFAULT": "0.25rem",
        "lg": "0.5rem",
        "xl": "0.75rem",
        "full": "9999px"
      },
    },
  },
  plugins: [],
}