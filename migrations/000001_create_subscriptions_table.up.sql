CREATE TABLE IF NOT EXISTS subscriptions (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    service    VARCHAR(100) NOT NULL,
    price      INTEGER      NOT NULL,
    user_id    UUID         NOT NULL,
    start_date DATE         NOT NULL,
    end_date   DATE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
