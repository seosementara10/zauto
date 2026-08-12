/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './layout.html',
    './pages/**/*.html',
    './app.js',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Segoe UI', 'system-ui', 'sans-serif'],
      },
      colors: {
        brand: {
          50: '#eff6ff',
          100: '#dbeafe',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          900: '#1e3a8a',
        },
      },
      boxShadow: {
        panel: '0 4px 24px rgba(15, 23, 42, 0.08)',
      },
    },
  },
  plugins: [],
};
