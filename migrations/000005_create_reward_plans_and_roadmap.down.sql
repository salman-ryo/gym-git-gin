-- Migration 000005 Down: Drop user_claimed_rewards, reward_plan_milestones, and reward_plans tables

DROP TABLE IF EXISTS user_claimed_rewards CASCADE;
DROP TABLE IF EXISTS reward_plan_milestones CASCADE;
DROP TABLE IF EXISTS reward_plans CASCADE;
