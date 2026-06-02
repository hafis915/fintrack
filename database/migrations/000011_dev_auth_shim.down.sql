-- We never drop auth.uid(): on Supabase it's not ours to drop, and locally
-- dropping it would break the RLS policies the next migration depends on.
-- This down is intentionally a no-op.
SELECT 1;
