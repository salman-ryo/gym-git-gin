-- Add checkin_snoozed_date and checkin_snoozed_at to users table for dynamic check-in snooze tracking
ALTER TABLE users ADD COLUMN IF NOT EXISTS checkin_snoozed_date VARCHAR(10);
ALTER TABLE users ADD COLUMN IF NOT EXISTS checkin_snoozed_at TIMESTAMPTZ;
