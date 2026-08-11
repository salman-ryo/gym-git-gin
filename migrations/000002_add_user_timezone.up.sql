-- Migration: Add timezone column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(100) NOT NULL DEFAULT 'UTC';
