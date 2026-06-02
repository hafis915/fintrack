CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE expense_category_type AS ENUM ('fixed','variable','debt','want');
CREATE TYPE financial_program     AS ENUM ('pondasi','bebas_utang','goal_chaser','tumbuh','seimbang');
CREATE TYPE fatigue_status        AS ENUM ('fresh','warning','fatigued');
CREATE TYPE debt_method           AS ENUM ('snowball','avalanche');
CREATE TYPE lifestyle_style       AS ENUM ('easy','balanced','strict');
CREATE TYPE housing_type          AS ENUM ('kosan','kpr','keluarga');
