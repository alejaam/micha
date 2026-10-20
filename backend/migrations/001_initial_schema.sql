-- 001_initial_schema.sql
-- Baseline schema for micha v1.
-- Defines all core tables, constraints, foreign keys, and indexes in dependency order.

-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id            TEXT        PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Households
CREATE TABLE IF NOT EXISTS households (
    id               TEXT        PRIMARY KEY,
    name             TEXT        NOT NULL,
    owner_id         TEXT        REFERENCES users(id) ON DELETE SET NULL,
    settlement_mode  TEXT        NOT NULL CHECK (settlement_mode IN ('equal', 'proportional')),
    currency         CHAR(3)     NOT NULL DEFAULT 'MXN',
    closing_day      INT         NOT NULL DEFAULT 15 CHECK (closing_day >= 1 AND closing_day <= 31),
    period_frequency TEXT        NOT NULL DEFAULT 'monthly' CHECK (period_frequency IN ('monthly', 'biweekly')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_households_owner_id ON households(owner_id);
CREATE INDEX IF NOT EXISTS idx_households_created_at ON households(created_at DESC);

-- 3. Members
CREATE TABLE IF NOT EXISTS members (
    id                   TEXT        PRIMARY KEY,
    household_id         TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    user_id              TEXT        REFERENCES users(id) ON DELETE SET NULL,
    name                 TEXT        NOT NULL,
    email                TEXT        NOT NULL,
    monthly_salary_cents BIGINT      NOT NULL CHECK (monthly_salary_cents >= 0),
    deleted_at           TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT members_household_id_id_key UNIQUE (household_id, id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_members_household_email_active ON members(household_id, email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_members_household_created ON members(household_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_members_user_id ON members(user_id) WHERE deleted_at IS NULL;

-- 4. Household Split Config
CREATE TABLE IF NOT EXISTS household_split_config (
    household_id TEXT         NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    member_id    TEXT         NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    percentage   NUMERIC(5,2) NOT NULL CHECK (percentage >= 0 AND percentage <= 100),
    PRIMARY KEY (household_id, member_id),
    CONSTRAINT fk_split_config_member_same_household
        FOREIGN KEY (household_id, member_id)
        REFERENCES members (household_id, id)
        ON DELETE CASCADE
);

-- 5. Member Invitations
CREATE TABLE IF NOT EXISTS member_invitations (
    id           TEXT        PRIMARY KEY,
    household_id TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    member_id    TEXT        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    email        TEXT        NOT NULL,
    invite_code  TEXT        NOT NULL UNIQUE,
    expires_at   TIMESTAMPTZ NOT NULL,
    used_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_member_invitations_household_email ON member_invitations(household_id, email, created_at DESC);

-- 6. Categories
CREATE TABLE IF NOT EXISTS categories (
    id           TEXT        PRIMARY KEY,
    household_id TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    slug         TEXT        NOT NULL,
    is_default   BOOLEAN     NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT categories_household_slug_key UNIQUE (household_id, slug),
    CONSTRAINT categories_household_id_id_key UNIQUE (household_id, id)
);

CREATE INDEX IF NOT EXISTS idx_categories_household_id ON categories(household_id);

-- 7. Cards
CREATE TABLE IF NOT EXISTS cards (
    id              TEXT        PRIMARY KEY,
    household_id    TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    owner_member_id TEXT        REFERENCES members(id) ON DELETE SET NULL,
    bank_name       TEXT        NOT NULL,
    card_name       TEXT        NOT NULL,
    cutoff_day      INT         NOT NULL CHECK (cutoff_day >= 1 AND cutoff_day <= 31),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT cards_household_id_id_key UNIQUE (household_id, id),
    CONSTRAINT fk_cards_owner_same_household
        FOREIGN KEY (household_id, owner_member_id)
        REFERENCES members (household_id, id)
        ON DELETE SET NULL (owner_member_id)
);

CREATE INDEX IF NOT EXISTS idx_cards_household ON cards(household_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cards_owner_member ON cards(owner_member_id) WHERE owner_member_id IS NOT NULL;

-- 8. Periods
CREATE TABLE IF NOT EXISTS periods (
    id           TEXT        PRIMARY KEY,
    household_id TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    start_date   TIMESTAMPTZ NOT NULL,
    end_date     TIMESTAMPTZ NOT NULL,
    status       TEXT        NOT NULL CHECK (status IN ('open', 'review', 'closed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT check_period_dates CHECK (end_date >= start_date),
    CONSTRAINT periods_household_id_id_key UNIQUE (household_id, id)
);

CREATE INDEX IF NOT EXISTS idx_periods_household_id ON periods(household_id);
CREATE INDEX IF NOT EXISTS idx_periods_status ON periods(status);

-- 9. Expenses
CREATE TABLE IF NOT EXISTS expenses (
    id                 TEXT        PRIMARY KEY,
    household_id       TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    paid_by_member_id  TEXT        REFERENCES members(id) ON DELETE SET NULL,
    period_id          TEXT        REFERENCES periods(id) ON DELETE SET NULL,
    category_id        TEXT        NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    card_id            TEXT        REFERENCES cards(id) ON DELETE SET NULL,
    amount_cents       BIGINT      NOT NULL CHECK (amount_cents > 0),
    description        TEXT        NOT NULL DEFAULT '',
    is_shared          BOOLEAN     NOT NULL DEFAULT true,
    currency           CHAR(3)     NOT NULL DEFAULT 'MXN',
    payment_method     TEXT        NOT NULL DEFAULT 'cash' CHECK (payment_method IN ('cash', 'card', 'transfer', 'voucher')),
    expense_type       TEXT        NOT NULL DEFAULT 'variable' CHECK (expense_type IN ('fixed', 'variable', 'msi')),
    card_name          TEXT        NOT NULL DEFAULT '',
    total_installments INT         NOT NULL DEFAULT 0 CHECK (total_installments >= 0),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,

    CONSTRAINT fk_expenses_member_same_household
        FOREIGN KEY (household_id, paid_by_member_id)
        REFERENCES members (household_id, id)
        ON DELETE SET NULL (paid_by_member_id),
    CONSTRAINT fk_expenses_category_same_household
        FOREIGN KEY (household_id, category_id)
        REFERENCES categories (household_id, id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_expenses_card_same_household
        FOREIGN KEY (household_id, card_id)
        REFERENCES cards (household_id, id)
        ON DELETE SET NULL (card_id),
    CONSTRAINT fk_expenses_period_same_household
        FOREIGN KEY (household_id, period_id)
        REFERENCES periods (household_id, id)
        ON DELETE SET NULL (period_id)
);

CREATE INDEX IF NOT EXISTS idx_expenses_household_created ON expenses(household_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_household_period ON expenses(household_id, created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_period_id ON expenses(period_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_card ON expenses(card_id) WHERE card_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_paid_by_member ON expenses(paid_by_member_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_household_shared ON expenses(household_id, is_shared, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_category_id ON expenses(category_id) WHERE deleted_at IS NULL;

-- 10. Installments (MSI)
CREATE TABLE IF NOT EXISTS installments (
    id                       TEXT        PRIMARY KEY,
    expense_id               TEXT        NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    paid_by_member_id        TEXT        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    start_date               DATE        NOT NULL,
    installment_amount_cents BIGINT      NOT NULL CHECK (installment_amount_cents > 0),
    total_amount_cents       BIGINT      NOT NULL CHECK (total_amount_cents > 0),
    total_installments       INT         NOT NULL CHECK (total_installments > 0),
    current_installment      INT         NOT NULL CHECK (current_installment > 0 AND current_installment <= total_installments),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_installments_start_date ON installments(start_date);
CREATE INDEX IF NOT EXISTS idx_installments_expense_id ON installments(expense_id);
CREATE INDEX IF NOT EXISTS idx_installments_paid_by_member ON installments(paid_by_member_id);
CREATE INDEX IF NOT EXISTS idx_installments_period_lookup ON installments(start_date, paid_by_member_id);

-- 11. Recurring Expenses
CREATE TABLE IF NOT EXISTS recurring_expenses (
    id                   TEXT        PRIMARY KEY,
    household_id         TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    paid_by_member_id    TEXT        REFERENCES members(id) ON DELETE CASCADE,
    amount_cents         BIGINT      NOT NULL CHECK (amount_cents > 0),
    description          TEXT        NOT NULL DEFAULT '',
    category_id          TEXT        NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    expense_type         TEXT        NOT NULL DEFAULT 'fixed' CHECK (expense_type IN ('fixed', 'variable', 'msi')),
    recurrence_pattern   TEXT        NOT NULL CHECK (recurrence_pattern IN ('monthly', 'biweekly', 'weekly')),
    start_date           DATE        NOT NULL,
    end_date             DATE,
    next_generation_date DATE        NOT NULL,
    is_active            BOOLEAN     NOT NULL DEFAULT true,
    is_agnostic          BOOLEAN     NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ,

    CONSTRAINT check_end_date_after_start
        CHECK (end_date IS NULL OR end_date >= start_date),
    CONSTRAINT check_recurring_expenses_agnostic_consistency
        CHECK (
            (is_agnostic = true AND expense_type = 'fixed' AND paid_by_member_id IS NULL)
            OR
            (is_agnostic = false AND paid_by_member_id IS NOT NULL AND btrim(paid_by_member_id) <> '')
        ),
    CONSTRAINT fk_recurring_member_same_household
        FOREIGN KEY (household_id, paid_by_member_id)
        REFERENCES members (household_id, id)
        ON DELETE CASCADE,
    CONSTRAINT fk_recurring_category_same_household
        FOREIGN KEY (household_id, category_id)
        REFERENCES categories (household_id, id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_recurring_expenses_household
    ON recurring_expenses (household_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_recurring_expenses_next_generation
    ON recurring_expenses (next_generation_date)
    WHERE deleted_at IS NULL AND is_active = true;

-- 12. Subscription Services Catalog
CREATE TABLE IF NOT EXISTS subscription_services (
    id                     TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name                   TEXT        NOT NULL,
    slug                   TEXT        NOT NULL UNIQUE,
    region                 TEXT        NOT NULL DEFAULT 'MX',
    currency               TEXT        NOT NULL DEFAULT 'MXN',
    standalone_price_cents BIGINT      NOT NULL CHECK (standalone_price_cents > 0),
    icon_url               TEXT,
    is_bundle              BOOLEAN     NOT NULL DEFAULT false,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscription_services_slug ON subscription_services (slug);

-- 13. Junction Table: Recurring Expense Catalog Links
CREATE TABLE IF NOT EXISTS recurring_expense_catalog_links (
    recurring_expense_id TEXT        NOT NULL REFERENCES recurring_expenses(id) ON DELETE CASCADE,
    catalog_service_id   TEXT        NOT NULL REFERENCES subscription_services(id) ON DELETE CASCADE,
    custom_price_cents   BIGINT      CHECK (custom_price_cents IS NULL OR custom_price_cents > 0),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (recurring_expense_id, catalog_service_id)
);

CREATE INDEX IF NOT EXISTS idx_catalog_links_service ON recurring_expense_catalog_links(catalog_service_id);
