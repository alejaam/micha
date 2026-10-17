-- Follow-up integrity constraints. Existing migrations are immutable; all
-- compatibility fixes are added here.

ALTER TABLE expenses
    ADD CONSTRAINT fk_expenses_household
        FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE
        NOT VALID;

ALTER TABLE households
    ADD CONSTRAINT fk_households_owner
        FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL
        NOT VALID;

ALTER TABLE periods
    ADD CONSTRAINT check_period_dates
        CHECK (end_date >= start_date)
        NOT VALID;

ALTER TABLE member_invitations
    ADD CONSTRAINT member_invitations_invite_code_key UNIQUE (invite_code);

ALTER TABLE recurring_expenses
    DROP CONSTRAINT fk_recurring_expenses_category,
    ADD CONSTRAINT fk_recurring_expenses_category
        FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
        NOT VALID;

ALTER TABLE members
    ADD CONSTRAINT members_household_id_id_key UNIQUE (household_id, id);

ALTER TABLE categories
    ADD CONSTRAINT categories_household_id_id_key UNIQUE (household_id, id);

ALTER TABLE cards
    ADD CONSTRAINT cards_household_id_id_key UNIQUE (household_id, id);

ALTER TABLE periods
    ADD CONSTRAINT periods_household_id_id_key UNIQUE (household_id, id);

ALTER TABLE expenses
    ADD CONSTRAINT fk_expenses_member_same_household
        FOREIGN KEY (household_id, paid_by_member_id)
        REFERENCES members (household_id, id)
        ON DELETE SET NULL (paid_by_member_id)
        NOT VALID,
    ADD CONSTRAINT fk_expenses_category_same_household
        FOREIGN KEY (household_id, category_id)
        REFERENCES categories (household_id, id)
        ON DELETE RESTRICT
        NOT VALID,
    ADD CONSTRAINT fk_expenses_card_same_household
        FOREIGN KEY (household_id, card_id)
        REFERENCES cards (household_id, id)
        ON DELETE SET NULL (card_id)
        NOT VALID,
    ADD CONSTRAINT fk_expenses_period_same_household
        FOREIGN KEY (household_id, period_id)
        REFERENCES periods (household_id, id)
        ON DELETE SET NULL (period_id)
        NOT VALID;

ALTER TABLE recurring_expenses
    ADD CONSTRAINT fk_recurring_member_same_household
        FOREIGN KEY (household_id, paid_by_member_id)
        REFERENCES members (household_id, id)
        ON DELETE CASCADE
        NOT VALID,
    ADD CONSTRAINT fk_recurring_category_same_household
        FOREIGN KEY (household_id, category_id)
        REFERENCES categories (household_id, id)
        ON DELETE RESTRICT
        NOT VALID;

ALTER TABLE cards
    ADD CONSTRAINT fk_cards_owner_same_household
        FOREIGN KEY (household_id, owner_member_id)
        REFERENCES members (household_id, id)
        ON DELETE SET NULL (owner_member_id)
        NOT VALID;
