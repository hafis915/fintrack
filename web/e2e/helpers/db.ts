import { execSync } from 'node:child_process'

// Wipes the onboarding-related tables in fintrack_test so each e2e spec
// starts from a known baseline. System default categories are NOT touched
// (they come from the migration seed and the API relies on them).
export function resetOnboardingDB() {
  const sql = [
    'delete from budget_items;',
    'delete from budget_plans;',
    'delete from user_profiles;',
    'delete from expense_categories where user_id is not null;',
    'delete from users;',
  ].join(' ')

  execSync(
    `docker exec fintrack-postgres psql -U fintrack -d fintrack_test -c "${sql}"`,
    { stdio: ['ignore', 'ignore', 'ignore'] },
  )
}
