-- Migration 000005: Create reward_plans, reward_plan_milestones, and user_claimed_rewards tables

CREATE TABLE IF NOT EXISTS reward_plans (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reward_plan_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id VARCHAR(50) NOT NULL REFERENCES reward_plans(id) ON DELETE CASCADE,
    streak_target INT NOT NULL,
    item_id VARCHAR(50) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    badge_slug VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(plan_id, streak_target, item_id)
);
CREATE INDEX IF NOT EXISTS idx_reward_plan_milestones_plan_target ON reward_plan_milestones(plan_id, streak_target);

CREATE TABLE IF NOT EXISTS user_claimed_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id VARCHAR(50) NOT NULL REFERENCES reward_plans(id) ON DELETE CASCADE,
    streak_target INT NOT NULL,
    item_id VARCHAR(50) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, plan_id, streak_target, item_id)
);
CREATE INDEX IF NOT EXISTS idx_user_claimed_rewards_user_plan ON user_claimed_rewards(user_id, plan_id);

-- Seed default reward plan
INSERT INTO reward_plans (id, name, description, is_active) VALUES
('default-streak-roadmap', 'Default Streak Progression Roadmap', 'Standard streak milestone rewards roadmap for Gym-Git users', TRUE)
ON CONFLICT (id) DO NOTHING;

-- Seed default reward plan milestones
INSERT INTO reward_plan_milestones (plan_id, streak_target, item_id, quantity, title, description, badge_slug) VALUES
('default-streak-roadmap', 7, 'RESTORE_SHIELD', 1, '7-Day Shield Anchor', 'Claim 1x Restore Shield to protect your consistency', 'shield-badge-bronze'),
('default-streak-roadmap', 10, 'STREAK_FREEZE_TOKEN', 1, '10-Day Ice Defender', 'Claim 1x Streak Freeze Token for sickness/rest days', 'freeze-badge-bronze'),
('default-streak-roadmap', 14, 'RESTORE_SHIELD', 1, '14-Day Iron Fortress', 'Claim 1x Restore Shield', 'shield-badge-silver'),
('default-streak-roadmap', 21, 'STREAK_FREEZE_TOKEN', 2, '21-Day Glacier Master', 'Claim 2x Streak Freeze Tokens', 'freeze-badge-silver'),
('default-streak-roadmap', 30, 'XP_BOOST', 1, '30-Day Power Surge', 'Claim 1x XP Boost (1.5x multiplier for 7 days)', 'xp-badge-gold'),
('default-streak-roadmap', 30, 'RESTORE_SHIELD', 1, '30-Day Power Surge Shield', 'Claim 1x Restore Shield', 'shield-badge-gold'),
('default-streak-roadmap', 60, 'STREAK_FREEZE_TOKEN', 2, '60-Day Titan Legion Tokens', 'Claim 2x Streak Freeze Tokens', 'freeze-badge-titan'),
('default-streak-roadmap', 60, 'RESTORE_SHIELD', 2, '60-Day Titan Legion Shields', 'Claim 2x Restore Shields', 'shield-badge-titan'),
('default-streak-roadmap', 90, 'RESTORE_SHIELD', 2, '90-Day Apex Champion Shields', 'Claim 2x Restore Shields', 'shield-badge-apex'),
('default-streak-roadmap', 90, 'XP_BOOST', 2, '90-Day Apex Champion Boosts', 'Claim 2x XP Boost Tokens', 'xp-badge-apex'),
('default-streak-roadmap', 180, 'STREAK_FREEZE_TOKEN', 3, '180-Day Mythic Legend Tokens', 'Claim 3x Streak Freeze Tokens', 'freeze-badge-mythic'),
('default-streak-roadmap', 180, 'RESTORE_SHIELD', 3, '180-Day Mythic Legend Shields', 'Claim 3x Restore Shields', 'shield-badge-mythic'),
('default-streak-roadmap', 365, 'STREAK_FREEZE_TOKEN', 5, '365-Day Immortal Demigod Tokens', 'Claim 5x Streak Freeze Tokens', 'freeze-badge-immortal'),
('default-streak-roadmap', 365, 'RESTORE_SHIELD', 5, '365-Day Immortal Demigod Shields', 'Claim 5x Restore Shields', 'shield-badge-immortal'),
('default-streak-roadmap', 365, 'XP_BOOST', 3, '365-Day Immortal Demigod Boosts', 'Claim 3x XP Boost Tokens', 'xp-badge-immortal')
ON CONFLICT (plan_id, streak_target, item_id) DO NOTHING;
