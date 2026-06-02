ALTER TABLE user_profiles      ENABLE ROW LEVEL SECURITY;
ALTER TABLE expense_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_plans       ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_items       ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions       ENABLE ROW LEVEL SECURITY;
ALTER TABLE debt_items         ENABLE ROW LEVEL SECURITY;
ALTER TABLE goals              ENABLE ROW LEVEL SECURITY;
ALTER TABLE weekly_reports     ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_tokens         ENABLE ROW LEVEL SECURITY;

CREATE POLICY "user owns user_profiles" ON user_profiles
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns budget_plans" ON budget_plans
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns budget_items via plan" ON budget_items
  FOR ALL USING (EXISTS (
    SELECT 1 FROM budget_plans bp WHERE bp.id = budget_plan_id AND bp.user_id = auth.uid()
  ));
CREATE POLICY "user owns transactions" ON transactions
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns debt_items" ON debt_items
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns goals" ON goals
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns weekly_reports" ON weekly_reports
  FOR ALL USING (user_id = auth.uid());
CREATE POLICY "user owns api_tokens" ON api_tokens
  FOR ALL USING (user_id = auth.uid());

CREATE POLICY "user reads system + own categories" ON expense_categories
  FOR SELECT USING (user_id IS NULL OR user_id = auth.uid());
CREATE POLICY "user inserts own categories" ON expense_categories
  FOR INSERT WITH CHECK (user_id = auth.uid());
CREATE POLICY "user updates own categories" ON expense_categories
  FOR UPDATE USING (user_id = auth.uid());
CREATE POLICY "user deletes own categories" ON expense_categories
  FOR DELETE USING (user_id = auth.uid());
