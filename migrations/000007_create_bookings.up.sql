CREATE TABLE bookings
(
    id                BIGSERIAL PRIMARY KEY,
    booking_code      VARCHAR(20)    NOT NULL,

    room_id           BIGINT         NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,

    guest_name        VARCHAR(255)   NOT NULL,
    guest_phone       VARCHAR(20)    NOT NULL,
    guest_email       VARCHAR(255),
    guest_note        TEXT,

    checkin_date      DATE           NOT NULL,
    checkout_date     DATE           NOT NULL,

    num_guests        INTEGER        NOT NULL DEFAULT 1 CHECK (num_guests > 0),
    num_nights        INTEGER        NOT NULL CHECK (num_nights > 0),

    base_total        DECIMAL(12, 0) NOT NULL DEFAULT 0 CHECK (base_total >= 0),
    cleaning_fee      DECIMAL(12, 0) NOT NULL DEFAULT 0 CHECK (cleaning_fee >= 0),
    discount          DECIMAL(12, 0) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    total_amount      DECIMAL(12, 0) NOT NULL CHECK (total_amount >= 0),
    currency          VARCHAR(10)    NOT NULL DEFAULT 'VND',

    price_breakdown   JSONB,

    status            booking_status NOT NULL DEFAULT 'pending',
    source            booking_source NOT NULL DEFAULT 'website',

    confirmed_at      TIMESTAMPTZ,
    canceled_at       TIMESTAMPTZ,
    cancel_reason     VARCHAR(255),
    expires_at        TIMESTAMPTZ,

    requires_refund   BOOLEAN        NOT NULL DEFAULT FALSE,
    refundable_amount DECIMAL(12, 0) NOT NULL DEFAULT 0,
    refund_status     refund_status  NOT NULL DEFAULT 'none',

    admin_note        TEXT,

    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_bookings_booking_code ON bookings (booking_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_bookings_guest_phone ON bookings (guest_phone);
CREATE INDEX idx_bookings_status ON bookings (status);
CREATE INDEX idx_bookings_refund_status ON bookings (refund_status);
CREATE INDEX idx_bookings_expires_at ON bookings (expires_at);
CREATE INDEX idx_bookings_room_date_range ON bookings (room_id, checkin_date, checkout_date);