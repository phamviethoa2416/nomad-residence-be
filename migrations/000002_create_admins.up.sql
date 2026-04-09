CREATE TABLE admins
(
    id                    BIGSERIAL PRIMARY KEY,
    email                 VARCHAR(255) NOT NULL,
    password_hash         VARCHAR(255) NOT NULL,
    full_name             VARCHAR(255) NOT NULL,
    phone                 VARCHAR(50),

    role                  admin_role   NOT NULL DEFAULT 'admin',
    status                VARCHAR(20)  NOT NULL DEFAULT 'active',

    failed_login_attempts INTEGER      NOT NULL DEFAULT 0,
    locked_until          TIMESTAMPTZ,
    last_login_at         TIMESTAMPTZ,
    password_changed_at   TIMESTAMPTZ,

    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_admins_email ON admins (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_admins_status ON admins (status);