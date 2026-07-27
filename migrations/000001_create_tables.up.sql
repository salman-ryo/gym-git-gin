-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: weekly_plans
CREATE TABLE IF NOT EXISTS weekly_plans (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    categories JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: users (App Profile mapping to Supabase Auth)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auth_user_id UUID UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(50) DEFAULT 'email',
    weekly_plan_id VARCHAR(50) REFERENCES weekly_plans(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: gym_logs
CREATE TABLE IF NOT EXISTS gym_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    hours NUMERIC(4,2) NOT NULL DEFAULT 0.00,
    workout_type VARCHAR(100) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for gym_logs
CREATE UNIQUE INDEX IF NOT EXISTS idx_gym_logs_user_date ON gym_logs (user_id, date);
CREATE INDEX IF NOT EXISTS idx_gym_logs_user_date_desc ON gym_logs (user_id, date DESC);

-- Seed initial weekly plans
INSERT INTO weekly_plans (id, name, description, categories) VALUES
('ppl', 'Push / Pull / Legs', '6-day high volume split focusing on Push, Pull, and Leg muscle groups.', '["Push", "Pull", "Legs", "Push", "Pull", "Legs"]'),
('upper_lower', 'Upper / Lower Split', '4-day balanced split alternating between upper and lower body workouts.', '["Upper", "Lower", "Upper", "Lower"]'),
('full_body', 'Full Body Split', '3-day efficiency split targeting the full body in every workout session.', '["Full Body", "Full Body", "Full Body"]')
ON CONFLICT (id) DO NOTHING;
