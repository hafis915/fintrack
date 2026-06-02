-- Dev-only shim: makes `auth.uid()` resolvable when running against vanilla
-- Postgres (Docker/brew/etc). On Supabase the `auth` schema and `auth.uid()`
-- already exist, so the IF NOT EXISTS guard makes this migration a no-op there
-- without overwriting Supabase's real implementation.
--
-- The local stub reads from a session GUC so that integration tests can
-- impersonate a user via `SET LOCAL request.jwt.claim.sub = '<uuid>'`.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_proc p
    JOIN pg_namespace n ON p.pronamespace = n.oid
    WHERE n.nspname = 'auth' AND p.proname = 'uid'
  ) THEN
    CREATE SCHEMA IF NOT EXISTS auth;
    EXECUTE $f$
      CREATE FUNCTION auth.uid() RETURNS uuid AS $body$
        SELECT NULLIF(current_setting('request.jwt.claim.sub', true), '')::uuid;
      $body$ LANGUAGE sql STABLE;
    $f$;
  END IF;
END
$$;
