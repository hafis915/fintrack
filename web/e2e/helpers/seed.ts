import { execSync } from 'node:child_process'

// Runs raw SQL inside the postgres container against fintrack_test. Used
// by e2e specs to seed budget plans + transactions so the UI has data to
// render without going through the full onboarding flow on every test.
export function exec(sql: string): string {
  const raw = execSync(
    `docker exec fintrack-postgres psql -U fintrack -d fintrack_test -tA -c "${sql.replace(/"/g, '\\"')}"`,
    { encoding: 'utf8' },
  )
  // psql prints the command tag ("INSERT 0 1") on a separate line even with
  // -tA. Strip empty + command-tag lines so callers get clean data.
  return raw
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l && !/^(INSERT|UPDATE|DELETE|SELECT)\s+\d+/.test(l))
    .join('\n')
}

// getDefaultCategoryIDs returns a map from system-default category name
// to its UUID. Tests can then pass real UUIDs into seedBudgetPlan / seedTx.
export function getDefaultCategoryIDs(): Record<string, string> {
  const out = exec(
    `select name || '|' || id from expense_categories where user_id is null`,
  )
  const map: Record<string, string> = {}
  for (const line of out.split('\n')) {
    const [name, id] = line.split('|')
    if (name && id) map[name] = id
  }
  return map
}

// seedUser inserts the auth shadow row so transactions/plans can FK to it.
export function seedUser(id: string): void {
  exec(`insert into users (id, email) values ('${id}', '${id}@local') on conflict do nothing`)
}

interface SeedPlanItem {
  categoryId: string
  allocated: number
  percentage: number
}

export function seedBudgetPlan(opts: {
  userId: string
  income: number
  program: string
  items: SeedPlanItem[]
  year?: number
  month?: number
}): string {
  const now = new Date()
  const year = opts.year ?? now.getFullYear()
  const month = opts.month ?? now.getMonth() + 1

  const planID = exec(
    `insert into budget_plans (user_id, period_year, period_month, total_income, program) ` +
      `values ('${opts.userId}', ${year}, ${month}, ${opts.income}, '${opts.program}') returning id`,
  )
  for (const it of opts.items) {
    exec(
      `insert into budget_items (budget_plan_id, category_id, allocated_amount, percentage) ` +
        `values ('${planID}', '${it.categoryId}', ${it.allocated}, ${it.percentage})`,
    )
  }
  return planID
}

export function seedTransaction(opts: {
  userId: string
  categoryId: string
  amount: number
  transactedAt?: string // RFC3339
}): void {
  const ts = opts.transactedAt ?? new Date().toISOString()
  exec(
    `insert into transactions (user_id, category_id, amount, transacted_at) ` +
      `values ('${opts.userId}', '${opts.categoryId}', ${opts.amount}, '${ts}')`,
  )
}
