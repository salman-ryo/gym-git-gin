-- Rollback: Drop user_streak_states table and queued_weekly_plan_id column

DROP TABLE IF EXISTS user_streak_states;
ALTER TABLE users DROP COLUMN IF EXISTS queued_weekly_plan_id;
