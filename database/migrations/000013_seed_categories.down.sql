DELETE FROM expense_categories WHERE user_id IS NULL AND is_default = true;
