-- Migration: Create items catalog, user_inventories, and user_active_effects tables

CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    effect_type VARCHAR(50) NOT NULL, -- 'INSTANT_USE', 'TIME_BASED'
    duration_seconds INT NOT NULL DEFAULT 0,
    rarity VARCHAR(30) NOT NULL DEFAULT 'common',
    icon_slug VARCHAR(100),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(50) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, item_id)
);
CREATE INDEX IF NOT EXISTS idx_user_inventories_user_id ON user_inventories(user_id);

CREATE TABLE IF NOT EXISTS user_active_effects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(50) NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_active_effects_user_expires ON user_active_effects(user_id, expires_at);

-- Seed core items catalog
INSERT INTO items (id, name, description, effect_type, duration_seconds, rarity, icon_slug, metadata) VALUES
('RESTORE_SHIELD', 'Restore Shield', 'Restores a lost streak from a missed day within the last 3 days.', 'INSTANT_USE', 0, 'rare', 'shield-icon', '{"lookback_days": 3}'),
('STREAK_FREEZE_TOKEN', 'Streak Freeze Token', 'Pauses streak decay for 24 hours per token without breaking active streak.', 'TIME_BASED', 86400, 'rare', 'snowflake-icon', '{"pause_reason": "Ice Pause"}'),
('XP_BOOST', 'XP Boost Token', 'Provides 1.5x multiplier to earned power points for 7 days.', 'TIME_BASED', 604800, 'epic', 'zap-icon', '{"multiplier": 1.5}'),
('ACCURACY_CHARM', 'Accuracy Charm', 'Protects workout split accuracy score from penalty for 1 cycle.', 'INSTANT_USE', 0, 'common', 'charm-icon', '{}')
ON CONFLICT (id) DO NOTHING;
