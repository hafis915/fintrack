-- Phase 2: receipt categorization.
-- Adds the merchant name extracted from a receipt photo (Claude Vision) so
-- the confirmed transaction can surface "where" the spend happened. Nullable:
-- manual entries without a receipt leave it NULL.

alter table transactions add column merchant text;
