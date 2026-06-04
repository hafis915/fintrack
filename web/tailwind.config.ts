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
          DEFAULT: '#F4A300',
          light: '#D97706',
        },
        // SPIKE: neo-brutalist light palette — broken-white bg, white cards,
        // black structural borders. (NOT Anthropic/Claude cream — see memory.)
        bg: '#F1F1EF',       // broken white / neutral light grey
        surface: '#FFFFFF',  // white cards (pop against grey + black border)
        elevated: '#E7E7E4', // grey for hover/elevated
        line: '#0A0A0A',     // black brutalist borders
        // Text
        fg: '#0A0A0A',       // near-black ink
        muted: '#52525B',    // neutral cool-grey, readable on broken white
        // Semantic — fatigue states (used as flat saturated blocks)
        fresh: '#16A34A',
        warning: '#F59E0B',
        fatigued: '#EF4444',
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
        // SPIKE: brutalism = sharp corners (was 12px).
        card: '2px',
      },
      boxShadow: {
        // SPIKE: signature hard offset brutalist shadow (no blur).
        brutal: '4px 4px 0 0 #0A0A0A',
        'brutal-sm': '2px 2px 0 0 #0A0A0A',
        'brutal-lg': '6px 6px 0 0 #0A0A0A',
      },
    },
  },
  plugins: [],
} satisfies Config
