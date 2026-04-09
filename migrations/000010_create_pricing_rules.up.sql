CREATE TABLE pricing_rules
(
    id             BIGSERIAL PRIMARY KEY,
    room_id        BIGINT            NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,

    name           VARCHAR(255),
    rule_type      pricing_rule_type NOT NULL,

    date_from      DATE,
    date_to        DATE,
    day_of_week    JSONB,

    price_modifier DECIMAL(12, 0)    NOT NULL,
    modifier_type  modifier_type     NOT NULL DEFAULT 'fixed',

    priority       INTEGER           NOT NULL DEFAULT 0,
    is_active      BOOLEAN           NOT NULL DEFAULT TRUE,

    created_at     TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_pricing_rules_room_id ON pricing_rules (room_id);
CREATE INDEX idx_pricing_rules_rule_type ON pricing_rules (rule_type);
CREATE INDEX idx_pricing_rules_date_range ON pricing_rules (date_from, date_to);

CREATE TABLE pricing_rule_days
(
    id      BIGSERIAL PRIMARY KEY,
    rule_id BIGINT   NOT NULL REFERENCES pricing_rules (id) ON DELETE CASCADE,
    day     SMALLINT NOT NULL CHECK (day >= 0 AND day <= 6)
);

CREATE INDEX idx_pricing_rule_days_rule_id ON pricing_rule_days (rule_id);