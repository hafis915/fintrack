---
title: "Fintrack — Design System"
type: project-doc
created: 2026-06-03
last_updated: 2026-06-03
tags: [project, design, design-system, fintrack]
related:
  - "[[(C) PROJECT.md]]"
  - "[[(C) ARCHITECTURE.md]]"
  - "[[(C) DECISIONS.md]]"
---

> **Always read this file before making any visual or UI decision.**
> Font choices, colors, spacing, motion — all defined here.
> Do not deviate without an ADR entry in `(C) DECISIONS.md`.

---

## Product Context

- **What:** Fintrack — personal finance PWA, mobile-first
- **Who:** Hafis (user #1). Indonesian fresh workers (Rp 8–10jt/mo) when public.
- **Space:** Personal finance / budgeting / fintech
- **Project type:** Mobile-first PWA, authenticated dashboard + tracking app

## The Memorable Thing (north star)

> **"Money discipline that feels like training, not bookkeeping."**

Every design decision below serves this single line. If a future choice doesn't reinforce this — it's wrong.

## Aesthetic Direction

- **Direction:** **Industrial Utilitarian** with editorial accents
- **Decoration level:** **Minimal** — typography and data carry the system
- **Mood:** Focused, disciplined, slightly stern. The product respects you enough to show the data straight.
- **Reference apps (steal from):**
  - [Whoop](https://www.whoop.com) — recovery score / strain rings (their Fresh/Warning/Fatigued analog)
  - [Strava](https://www.strava.com) — data-forward, bold numbers, athletic energy
  - [Apple Fitness+](https://www.apple.com/apple-fitness-plus/) — rings, vertical card stacks, dark mode as identity
  - [Strong](https://www.strong.app) / [Hevy](https://www.hevyapp.com) — utilitarian workout loggers, numbers as heroes
- **Anti-references (do NOT steal from):**
  - Cash App, Revolut, N26 (too playful/glassy)
  - Mercury, Brex (too corporate)
  - Traditional banking apps (too "trust us with your money" — Fintrack is "track yourself")

---

## Typography

### Font Stack

| Role | Font | Weight | Loading |
|------|------|--------|---------|
| **Hero numbers** ⭐ | **JetBrains Mono** | 700 (Bold), tabular | Google Fonts |
| **Display / Section headers** | **General Sans** | 700 (Bold), 600 (Semibold) | Fontshare |
| **Body** | **DM Sans** | 400 (Regular), 500 (Medium) | Google Fonts |
| **Labels / UI** | **DM Sans** | 500 (Medium) | Same as body |
| **Inline numbers in body** | DM Sans with `font-feature-settings: "tnum"` | — | Tabular figures setting |

### Hero Number Treatment (signature element)

The hero numbers are the brand. Not just "big bold digits" — they're a **designed composition**:

```
   Rp  2.450.000        ← Hero rendering
   ──  ─────────
   ↑       ↑
small/    huge/
saffron   white/JBM
500wt     700wt mono
```

**CSS sketch:**
```css
.hero-amount {
  font-family: "JetBrains Mono", monospace;
  font-weight: 700;
  font-size: 48px;
  line-height: 56px;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.hero-amount .currency {
  font-family: "General Sans", sans-serif;
  font-weight: 500;
  font-size: 24px;
  color: var(--brand-accent);  /* saffron */
  margin-right: 8px;
  vertical-align: baseline;
}

.hero-amount .decimals {
  color: var(--text-muted);
  font-size: 0.7em;
}
```

### Type Scale (mobile-first, 4px base)

| Level | Size / Line | Use |
|-------|-------------|-----|
| Hero number | `48px / 56px` | Dashboard primary stat (current balance, monthly spend) |
| Display | `32px / 40px` | Page titles, onboarding hero |
| H1 | `24px / 32px` | Card titles |
| H2 | `20px / 28px` | Section headers |
| Body large | `17px / 26px` | Important paragraphs |
| Body | `15px / 24px` | Default body |
| Caption | `13px / 18px` | Metadata, timestamps |
| Micro | `11px / 14px` | Labels, badges |

### Number Formatting Rules

- **Use tabular-nums everywhere numbers appear** — `font-variant-numeric: tabular-nums`
- **Indonesian format:** thousands separator `.` (period), decimals `,` (comma) → `Rp 2.450.000,50`
- **Currency prefix:** always `Rp ` with non-breaking space
- **Negative amounts:** prepend `-`, color `--fatigued` (coral red)
- **Decimals optional:** show only when meaningful (typically hide in main app, show in transaction detail)

---

## Color System

Dark mode is the default. Light mode is opt-in. **Both must be tested for every component.**

### Dark Mode (default)

```css
:root[data-theme="dark"] {
  /* Surfaces */
  --bg: #0A0A0B;
  --surface: #1A1A1C;
  --surface-elevated: #252527;
  --border: #2E2E30;
  --border-strong: #3E3E40;

  /* Text */
  --text-primary: #FFFFFF;
  --text-secondary: #A1A1A6;
  --text-muted: #6B6B70;
  --text-disabled: #4A4A4E;

  /* Brand */
  --brand-accent: #F4A300;        /* Saffron Gold — the signature */
  --brand-accent-hover: #FFB924;
  --brand-accent-muted: #8B5E00;  /* For backgrounds, hover surfaces */

  /* Semantic — Fatigue Dashboard core */
  --fresh: #34D399;
  --fresh-bg: rgba(52, 211, 153, 0.12);
  --warning: #FBBF24;
  --warning-bg: rgba(251, 191, 36, 0.12);
  --fatigued: #F87171;
  --fatigued-bg: rgba(248, 113, 113, 0.12);
  --info: #60A5FA;
  --info-bg: rgba(96, 165, 250, 0.12);
}
```

### Light Mode

```css
:root[data-theme="light"] {
  --bg: #FAFAFA;
  --surface: #FFFFFF;
  --surface-elevated: #F5F5F7;
  --border: #E5E5EA;
  --border-strong: #D1D1D6;

  --text-primary: #0A0A0B;
  --text-secondary: #6B6B70;
  --text-muted: #A1A1A6;
  --text-disabled: #C7C7CC;

  --brand-accent: #D97706;        /* Deeper saffron for white-bg contrast */
  --brand-accent-hover: #B45309;
  --brand-accent-muted: #FEF3C7;

  --fresh: #10B981;
  --fresh-bg: #D1FAE5;
  --warning: #D97706;
  --warning-bg: #FEF3C7;
  --fatigued: #DC2626;
  --fatigued-bg: #FEE2E2;
  --info: #2563EB;
  --info-bg: #DBEAFE;
}
```

### Color Usage Rules

1. **Saffron is sacred.** Use only for: the "Rp" currency prefix, primary CTAs, active nav tab, brand markers. Never decorative.
2. **Semantic colors are for STATE, not decoration.** Fresh/Warning/Fatigued are tied to fatigue dashboard logic — don't use them for buttons or links.
3. **Background hierarchy:** `--bg` < `--surface` < `--surface-elevated`. Cards lift visually by going lighter (dark mode) or by shadow (light mode).
4. **Contrast minimums:** WCAG AA for all text. Test in both modes.
5. **No gradients** anywhere except the receipt scan progress shimmer.

---

## Spacing

### Scale (4px base)

| Token | Value | Use |
|-------|-------|-----|
| `2xs` | `4px` | Tight inline gaps |
| `xs` | `8px` | Compact element spacing |
| `sm` | `12px` | Default form spacing |
| `md` | `16px` | Standard card padding |
| `lg` | `24px` | Section spacing |
| `xl` | `32px` | Major section gaps |
| `2xl` | `48px` | Page section dividers |
| `3xl` | `64px` | Hero / onboarding screens |

### Density

- **Compact.** Mobile screens are precious — no generous SaaS-marketing spacing.
- **Touch targets:** 44px minimum (Apple HIG) — applies to all tappable elements.
- **Thumb-zone aware:** primary actions in the bottom 1/3 of the viewport when possible. Bottom tab nav + floating action button (FAB) for receipt scan.
- **Comfortable line-height** despite tight spacing — readability stays priority.

---

## Layout

### Primary Patterns

- **Mobile-first single column.** Max content width `420px` even on desktop — this is a phone app.
- **Vertical card stacks** for dashboard (Apple Fitness pattern).
- **Bottom tab navigation** (5 tabs):
  1. Dashboard (home)
  2. Transactions (list)
  3. **Scan** (center, raised FAB style — hero action)
  4. Reports
  5. Profile
- **No sidebars, no top nav menus.** Mobile-first means thumb-first.

### Border Radius Hierarchy

| Token | Value | Use |
|-------|-------|-----|
| `radius-sm` | `8px` | Buttons, inputs, badges, small chips |
| `radius-md` | `16px` | Main content cards (fatigue cards, transactions) |
| `radius-lg` | `24px` | Modals, bottom sheets, onboarding cards |
| `radius-full` | `9999px` | Avatars, status pills, FAB, rings |

### Grid

- **Mobile:** Single column, 16px horizontal padding from screen edge
- **Tablet/Desktop:** Single column max-width `420px`, centered, 64px from screen edge
- **No multi-column dashboards** — vertical scroll only

---

## Motion (DYNAMIC — the signature)

> Motion is not decoration here. It's how Fintrack feels alive.

### Approach

**Intentional, leaning expressive on the data layer.** Numbers count. States transition visibly. The product responds.

### Signature Motions

| Element | Motion |
|---------|--------|
| **Hero numbers (on load + update)** | Count-up animation 0 → final value over 600ms, ease-out. Tick animation on every update (single digit roll if 3-digit change, otherwise just count). **This is THE signature motion.** |
| **Fatigue state card (status change)** | Background color cross-fades over 400ms, ring pulses once on change (scale 1.0 → 1.05 → 1.0 over 600ms spring). |
| **Page transitions** | Horizontal slide + fade, 250ms ease-out. iOS-native feel. |
| **Receipt scan flow** | Camera preview → capture (subtle flash) → scan shimmer (saffron gradient sweeps across image) → results materialize (fields stagger in, 50ms each, ease-out). |
| **Button press** | Scale 0.96 + opacity 0.8 for 100ms, spring back. |
| **Pull-to-refresh** | Native iOS-like elastic stretch with saffron loading ring. |
| **Tab switch (bottom nav)** | Active tab icon scales 1.0 → 1.15 → 1.0 over 200ms spring, accent color fades in. |
| **Modal / Sheet** | Slide up from bottom with backdrop fade, 300ms ease-out. iOS sheet feel. |

### Easing Curves

```css
--ease-out: cubic-bezier(0.16, 1, 0.3, 1);     /* Enters */
--ease-in: cubic-bezier(0.7, 0, 0.84, 0);      /* Exits */
--ease-in-out: cubic-bezier(0.83, 0, 0.17, 1); /* Moves */
--ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1); /* Bounces */
```

### Duration Scale

| Token | Value | Use |
|-------|-------|-----|
| `dur-micro` | `100ms` | Button presses, hover states |
| `dur-short` | `250ms` | Tab switches, small transitions |
| `dur-medium` | `400ms` | State changes (fatigue card colors) |
| `dur-long` | `600ms` | Hero count-ups, big reveals |
| `dur-xlong` | `800ms` | Onboarding entrance animations |

### Implementation

- **Library:** Vue 3 `<Transition>` + `@formkit/auto-animate` for list reorders
- **For count-up:** custom composable `useCountUp(targetValue, duration)` — no library needed for simple animation
- **For complex sequences (receipt scan flow):** `@vueuse/motion` or hand-rolled with Web Animations API
- **Avoid:** GSAP unless we genuinely need timeline orchestration. Adds 50KB+ to bundle.

### Motion Discipline

- **No purely decorative animations.** Every motion must communicate state, hierarchy, or progress.
- **Respect `prefers-reduced-motion`** — disable count-ups, fades, springs; keep functional transitions only.
- **No infinite loops** (no perpetual spinners) — use determinate progress where possible.

---

## Component Patterns (high-level rules)

### Buttons

- **Primary:** Saffron Gold background, white text, `radius-sm`, 44px height, weight 600
- **Secondary:** Surface background, primary text, `radius-sm`, 1px border
- **Ghost:** Transparent, primary text, no border
- **Destructive:** Fatigued (coral) background, white text
- **Sizes:** Default 44px (touch target), Small 36px (only in dense contexts)
- **Disabled state:** opacity 0.4, no pointer

### Cards (Fatigue Dashboard core component)

- Surface background, `radius-md` (16px), `md` padding (16px)
- **Status accent:** 4px left border or top-strip in semantic color (fresh/warning/fatigued)
- **Heading:** H2 (20px) primary text
- **Status pill:** `radius-full`, semantic background tint, semantic text color, 11px micro
- **Hero number:** large amount with internal hierarchy
- **Coaching message:** body, secondary text color

### Forms (Onboarding 6-step + transaction entry)

- Inputs: surface background, `radius-sm`, 48px height, body large text inside
- Labels: micro (11px) above input, muted color, uppercase letter-spacing 0.05em
- Focus state: 2px saffron ring (outline-offset 2px)
- Error state: 2px fatigued border + fatigued text below
- Step indicator: progress dots, saffron for completed, border-strong for upcoming

---

## Iconography

- **Library:** [Lucide](https://lucide.dev) — clean, consistent, free, 1.5px stroke
- **Sizes:** 16px (inline), 20px (default UI), 24px (nav), 32px (large headers)
- **Color:** inherit text color by default. Saffron only for active states.
- **No emoji** in UI chrome (allowed in user-generated transaction notes only).

---

## Accessibility (non-negotiable)

- All text meets WCAG AA contrast minimums in both themes
- Touch targets ≥ 44×44px
- Focus indicators visible on all interactive elements (2px saffron ring)
- `prefers-reduced-motion` respected
- Screen reader labels on icon-only buttons (use `aria-label`)
- Form fields have associated `<label>` elements
- Error messages programmatically associated with inputs

---

## Implementation Notes (Tailwind + shadcn-vue)

### Tailwind Config

```ts
// tailwind.config.ts
export default {
  darkMode: ["class", "[data-theme='dark']"],
  theme: {
    extend: {
      colors: {
        bg: "var(--bg)",
        surface: "var(--surface)",
        "surface-elevated": "var(--surface-elevated)",
        border: "var(--border)",
        "text-primary": "var(--text-primary)",
        "text-secondary": "var(--text-secondary)",
        "text-muted": "var(--text-muted)",
        accent: "var(--brand-accent)",
        fresh: "var(--fresh)",
        warning: "var(--warning)",
        fatigued: "var(--fatigued)",
      },
      fontFamily: {
        sans: ["DM Sans", "system-ui", "sans-serif"],
        display: ["General Sans", "DM Sans", "sans-serif"],
        mono: ["JetBrains Mono", "monospace"],
      },
      borderRadius: {
        sm: "8px",
        md: "16px",
        lg: "24px",
      },
      spacing: {
        "2xs": "4px",
        xs: "8px",
        sm: "12px",
        md: "16px",
        lg: "24px",
        xl: "32px",
        "2xl": "48px",
        "3xl": "64px",
      },
    },
  },
}
```

### Font Loading Strategy

In `web/index.html`:
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="preconnect" href="https://api.fontshare.com" crossorigin>

<!-- DM Sans + JetBrains Mono from Google Fonts -->
<link href="https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;700&family=JetBrains+Mono:wght@500;700&display=swap" rel="stylesheet">

<!-- General Sans from Fontshare -->
<link href="https://api.fontshare.com/v2/css?f[]=general-sans@500,600,700&display=swap" rel="stylesheet">
```

For PWA offline: consider self-hosting fonts in `web/public/fonts/` once the choice is final. Defer until v1 ships.

### Theme Toggle

- Toggle lives in Profile screen
- Saves to `user_profile.theme` column (`dark` | `light`)
- On app boot: apply `data-theme` attribute to `<html>` based on stored value
- Default for new users: `dark`

---

## Decisions Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-06-03 | Memorable thing: "Money discipline that feels like training, not bookkeeping." | Hafis chose Direction A (Performance) from three options. Pulls from Whoop, Strava, Apple Fitness rather than fintech. |
| 2026-06-03 | Dark mode default, light mode opt-in | Performance/training aesthetic favors dark. Light mode added per user request for choice. |
| 2026-06-03 | Saffron Gold (#F4A300) as brand accent | Indonesian cultural resonance, warm, distinct from green/blue fintech default. |
| 2026-06-03 | JetBrains Mono for hero numbers | Hafis confirmed after considering taste concerns. Layered with typographic hierarchy treatment (saffron Rp prefix, muted decimals) to avoid AI-coded aesthetic. |
| 2026-06-03 | General Sans for display, DM Sans for body | Avoids overused Inter/Space Grotesk/Roboto. Both supported by free font services. |
| 2026-06-03 | Lucide for icons | Industry standard, clean, free, consistent with utilitarian aesthetic. |
| 2026-06-03 | Bottom tab nav with raised center FAB for Scan | Mobile-first, thumb-zone aware, hero feature gets the prime spot. |

---

## Connections

- [[(C) PROJECT.md]] — Project overview, MVP scope
- [[(C) ARCHITECTURE.md]] — Implementation stack (Vue 3 + Vite + Tailwind + shadcn-vue)
- [[(C) DECISIONS.md]] — Full ADR log including ADR-013 (design system)
- [[(C) ROADMAP.md]] — When each design element gets implemented (mostly Phase 1+ for forms, Phase 4 for fatigue cards)
