-- 0001_create_subscriptions.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
                                             id SERIAL PRIMARY KEY,
                                             service_name TEXT NOT NULL,
                                             price INTEGER NOT NULL,
                                             user_id UUID NOT NULL,
                                             start_date DATE NOT NULL,
                                             end_date DATE,
                                             created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions (service_name);
