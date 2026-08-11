-- Migration: Add queued_weekly_plan_id to users table and create user_streak_states table

ALTER TABLE users ADD COLUMN IF NOT EXISTS queued_weekly_plan_id VARCHAR(50) REFERENCES weekly_plans(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS user_streak_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_streak INT NOT NULL DEFAULT 0,
    longest_streak INT NOT NULL DEFAULT 0,
    cycle_start_date DATE NOT NULL,
    cycle_end_date DATE NOT NULL,
    workouts_completed_in_cycle INT NOT NULL DEFAULT 0,
    workouts_target_in_cycle INT NOT NULL DEFAULT 4,
    rest_tokens_total INT NOT NULL DEFAULT 3,
    rest_tokens_used INT NOT NULL DEFAULT 0,
    accuracy_score INT NOT NULL DEFAULT 100,
    last_logged_date DATE,
    is_frozen BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_streak_states_user_id ON user_streak_states(user_id);
