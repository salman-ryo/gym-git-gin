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
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    weekly_plan_id VARCHAR(50) REFERENCES weekly_plans(id) ON DELETE SET NULL,
    queued_weekly_plan_id VARCHAR(50) REFERENCES weekly_plans(id) ON DELETE SET NULL,
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

-- Seed initial weekly plans matching frontend options
INSERT INTO weekly_plans (id, name, description, categories) VALUES
('ppl-standard', 'Push / Pull / Legs (PPL)', 'Classic 4-day active split focusing on movement patterns.', '["Push", "Pull", "Legs", "Cardio", "Custom"]'),
('ppl-core', 'PPL + Core & Cardio', 'Comprehensive 5-day athletic split.', '["Push", "Pull", "Legs", "Core", "Cardio", "Custom"]'),
('upper-lower', 'Upper / Lower Split', '4-day hypertrophy split split into upper & lower body.', '["Upper Body", "Lower Body", "Core & Cardio", "Custom"]'),
('full-body', 'Full Body & Functional', '3-day full body strength & conditioning plan.', '["Full Body", "Cardio", "Mobility", "Custom"]')
ON CONFLICT (id) DO NOTHING;

-- Also ensure legacy ppl exists
INSERT INTO weekly_plans (id, name, description, categories) VALUES
('ppl', 'Push / Pull / Legs', '6-day high volume split focusing on Push, Pull, and Leg muscle groups.', '["Push", "Pull", "Legs", "Push", "Pull", "Legs"]')
ON CONFLICT (id) DO NOTHING;

-- Add user_id column to weekly_plans for custom user-created plans
ALTER TABLE weekly_plans ADD COLUMN IF NOT EXISTS user_id UUID;

-- Enable FK reference back to users table safely
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_weekly_plans_user'
    ) THEN
        ALTER TABLE weekly_plans ADD CONSTRAINT fk_weekly_plans_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_weekly_plans_user_id ON weekly_plans(user_id);
