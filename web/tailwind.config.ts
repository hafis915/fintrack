import type { Config } from 'tailwindcss'

// Design tokens — see DESIGN.md.
// "Money discipline that feels like training, not bookkeeping."
// Saffron accent is sacred (only currency prefix, primary CTAs, active states).
// Semantic colors are state-only, never decoration.
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx,js,jsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Brand
        saffron: {
          DEFAULT: '#F4A300', // dark mode primary
          light: '#D97706',   // light mode primary
        },
        // Surfaces (dark default)
        bg: '#0A0A0A',
        surface: '#141414',
        elevated: '#1F1F1F',
        line: '#2A2A2A',
        // Text
        fg: '#FAFAFA',
        muted: '#A1A1AA',
        // Semantic — fatigue states
        fresh: '#22C55E',
        warning: '#F59E0B',
        fatigued: '#F87171',
      },
      fontFamily: {
        // Hero numbers — JetBrains Mono (load via CSS/Google Fonts in Phase 2)
        mono: ['"JetBrains Mono"', 'ui-monospace', 'monospace'],
        // Display — General Sans (Phase 2 will add the @font-face)
        display: ['"General Sans"', 'system-ui', 'sans-serif'],
        // Body — DM Sans
        sans: ['"DM Sans"', 'system-ui', 'sans-serif'],
      },
      maxWidth: {
        mobile: '420px',
      },
      borderRadius: {
        card: '12px',
      },
    },
  },
  plugins: [],
} satisfies Config
