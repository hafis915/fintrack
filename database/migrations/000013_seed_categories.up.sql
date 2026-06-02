INSERT INTO expense_categories (user_id, name, icon, type, is_default, sort_order) VALUES
  (NULL, 'Sewa kosan',     '🏠', 'fixed',    true, 10),
  (NULL, 'Cicilan KPR',    '🏗', 'fixed',    true, 11),
  (NULL, 'Listrik & air',  '💡', 'fixed',    true, 12),
  (NULL, 'Transportasi',   '🛵', 'fixed',    true, 13),
  (NULL, 'Internet & HP',  '📶', 'fixed',    true, 14),
  (NULL, 'Makan & minum',  '🍱', 'variable', true, 20),
  (NULL, 'Belanja harian', '🛒', 'variable', true, 21),
  (NULL, 'Hiburan',        '🎬', 'want',     true, 30),
  (NULL, 'Self-care',      '💅', 'want',     true, 31),
  (NULL, 'Kartu kredit',   '💳', 'debt',     true, 40),
  (NULL, 'Paylater',       '💸', 'debt',     true, 41),
  (NULL, 'Tabungan',       '🐷', 'fixed',    true, 50);
