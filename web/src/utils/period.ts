// Month-period helpers for the reports view. Month filtering is a frontend
// concern: listTransactions({ from, to }) takes ISO date strings, so we compute
// a half-open [from, to) range covering exactly one calendar month.

export interface MonthRange {
  from: string // inclusive — first day of month, 00:00 local
  to: string // exclusive — first day of the next month, 00:00 local
}

// month is 1-based (1 = January, 12 = December), matching the friendly label.
// We build the bounds in local time then serialize to ISO so the API receives
// an unambiguous instant; the backend compares against transacted_at.
export function monthRange(year: number, month: number): MonthRange {
  const from = new Date(year, month - 1, 1, 0, 0, 0, 0)
  const to = new Date(year, month, 1, 0, 0, 0, 0) // month overflow rolls the year
  return { from: from.toISOString(), to: to.toISOString() }
}

const MONTH_LABELS_ID = [
  'Januari',
  'Februari',
  'Maret',
  'April',
  'Mei',
  'Juni',
  'Juli',
  'Agustus',
  'September',
  'Oktober',
  'November',
  'Desember',
]

// e.g. monthLabel(2026, 6) -> "Juni 2026"
export function monthLabel(year: number, month: number): string {
  return `${MONTH_LABELS_ID[month - 1]} ${year}`
}

// Used for the CSV filename: monthSlug(2026, 6) -> "2026-06".
export function monthSlug(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, '0')}`
}

export interface YearMonth {
  year: number
  month: number // 1-based
}

export function currentYearMonth(now: Date = new Date()): YearMonth {
  return { year: now.getFullYear(), month: now.getMonth() + 1 }
}

// Step a {year, month} by ±n months, normalizing year rollover.
export function shiftMonth(ym: YearMonth, delta: number): YearMonth {
  // Convert to a 0-based absolute month index, shift, convert back.
  const abs = ym.year * 12 + (ym.month - 1) + delta
  return { year: Math.floor(abs / 12), month: (((abs % 12) + 12) % 12) + 1 }
}

// A list of the most recent `count` months (newest first), for a dropdown.
export function recentMonths(count: number, now: Date = new Date()): YearMonth[] {
  const start = currentYearMonth(now)
  const out: YearMonth[] = []
  for (let i = 0; i < count; i++) out.push(shiftMonth(start, -i))
  return out
}
