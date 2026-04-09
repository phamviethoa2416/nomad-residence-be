CREATE TABLE payments
(
    id                      BIGSERIAL PRIMARY KEY,
    booking_id              BIGINT         NOT NULL REFERENCES bookings (id) ON DELETE CASCADE,

    amount                  DECIMAL(12, 0) NOT NULL CHECK (amount >= 0),
    currency                VARCHAR(10)    NOT NULL DEFAULT 'VND',
    method                  payment_method NOT NULL,
    status                  payment_status NOT NULL DEFAULT 'pending',

    idempotency_key         VARCHAR(100)   NOT NULL,
    external_transaction_id VARCHAR(100),

    paid_at                 TIMESTAMPTZ,
    raw_response            JSONB,
    admin_note              TEXT,

    created_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_payments_idempotency_key ON payments (idempotency_key);
CREATE UNIQUE INDEX idx_payments_external_transaction_id ON payments (external_transaction_id) WHERE external_transaction_id IS NOT NULL;
CREATE INDEX idx_payments_booking_id ON payments (booking_id);
CREATE INDEX idx_payments_method ON payments (method);
CREATE INDEX idx_payments_status ON payments (status);