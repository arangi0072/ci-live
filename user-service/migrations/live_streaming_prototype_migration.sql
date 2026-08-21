-- Live Streaming Platform Prototype PostgreSQL Migration
-- Auth Service section + Streaming Service section.
-- Run each section in its respective database.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ========================= AUTH SERVICE =========================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    device_id VARCHAR(255),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);

-- ====================== STREAMING SERVICE =======================
-- user_id values below are owned by Auth Service; cross-database FKs
-- are intentionally not used.

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    username VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    avatar_url TEXT,
    banner_url TEXT,
    follower_count BIGINT NOT NULL DEFAULT 0,
    stream_count BIGINT NOT NULL DEFAULT 0,
    is_live BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_channels_user_id ON channels(user_id);

CREATE TABLE IF NOT EXISTS channel_follows (
    channel_id UUID NOT NULL,
    user_id UUID NOT NULL,
    notification_level VARCHAR(20) NOT NULL DEFAULT 'ALL',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id),
    CONSTRAINT fk_channel_follows_channel FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    CONSTRAINT chk_channel_follow_notification_level CHECK (notification_level IN ('ALL', 'NONE'))
);
CREATE INDEX IF NOT EXISTS idx_channel_follows_user_id ON channel_follows(user_id);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon_url TEXT,
    cover_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    follower_count BIGINT NOT NULL DEFAULT 0,
    stream_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    category_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    visibility VARCHAR(20) NOT NULL DEFAULT 'PUBLIC',
    viewer_count INTEGER NOT NULL DEFAULT 0,
    peak_viewers INTEGER NOT NULL DEFAULT 0,
    like_count BIGINT NOT NULL DEFAULT 0,
    total_views BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_streams_channel FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    CONSTRAINT fk_streams_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    CONSTRAINT chk_stream_status CHECK (status IN ('CREATED','STARTING','LIVE','RECONNECTING','ENDING','ENDED','FAILED','CANCELLED')),
    CONSTRAINT chk_stream_visibility CHECK (visibility IN ('PUBLIC','UNLISTED','PRIVATE')),
    CONSTRAINT chk_stream_viewer_count CHECK (viewer_count >= 0),
    CONSTRAINT chk_stream_peak_viewers CHECK (peak_viewers >= 0),
    CONSTRAINT chk_stream_like_count CHECK (like_count >= 0),
    CONSTRAINT chk_stream_total_views CHECK (total_views >= 0)
);
CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);
CREATE INDEX IF NOT EXISTS idx_streams_live ON streams(status, created_at DESC) WHERE status = 'LIVE';
CREATE INDEX IF NOT EXISTS idx_streams_channel ON streams(channel_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_streams_category ON streams(category_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_streams_viewers ON streams(viewer_count DESC) WHERE status = 'LIVE';

CREATE TABLE IF NOT EXISTS stream_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    media_server_id VARCHAR(100),
    media_path VARCHAR(255) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    duration_seconds BIGINT,
    peak_viewers INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_stream_sessions_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE,
    CONSTRAINT chk_stream_session_duration CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
    CONSTRAINT chk_stream_session_peak_viewers CHECK (peak_viewers >= 0)
);
CREATE INDEX IF NOT EXISTS idx_stream_sessions_stream ON stream_sessions(stream_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_stream_sessions_media_path ON stream_sessions(media_path);

CREATE TABLE IF NOT EXISTS stream_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    endpoint_type VARCHAR(20) NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    url TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_stream_endpoints_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE,
    CONSTRAINT chk_stream_endpoint_type CHECK (endpoint_type IN ('PUBLISH','PLAYBACK')),
    CONSTRAINT chk_stream_endpoint_protocol CHECK (protocol IN ('RTMP','HLS')),
    UNIQUE (stream_id, endpoint_type, protocol)
);
CREATE INDEX IF NOT EXISTS idx_stream_endpoints_stream ON stream_endpoints(stream_id);

CREATE TABLE IF NOT EXISTS stream_recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    session_id UUID,
    storage_provider VARCHAR(20) NOT NULL DEFAULT 'MINIO',
    storage_key TEXT NOT NULL,
    playback_url TEXT,
    thumbnail_url TEXT,
    format VARCHAR(20),
    width INTEGER,
    height INTEGER,
    file_size BIGINT,
    duration_seconds BIGINT,
    view_count BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'PROCESSING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT fk_stream_recordings_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE,
    CONSTRAINT fk_stream_recordings_session FOREIGN KEY (session_id) REFERENCES stream_sessions(id) ON DELETE SET NULL,
    CONSTRAINT chk_recording_storage_provider CHECK (storage_provider IN ('MINIO')),
    CONSTRAINT chk_recording_status CHECK (status IN ('PROCESSING','READY','FAILED','DELETED')),
    CONSTRAINT chk_recording_dimensions CHECK ((width IS NULL AND height IS NULL) OR (width > 0 AND height > 0)),
    CONSTRAINT chk_recording_file_size CHECK (file_size IS NULL OR file_size >= 0),
    CONSTRAINT chk_recording_duration CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
    CONSTRAINT chk_recording_view_count CHECK (view_count >= 0)
);
CREATE INDEX IF NOT EXISTS idx_stream_recordings_stream ON stream_recordings(stream_id);
CREATE INDEX IF NOT EXISTS idx_stream_recordings_status ON stream_recordings(status);
CREATE INDEX IF NOT EXISTS idx_stream_recordings_created ON stream_recordings(created_at DESC);

CREATE TABLE IF NOT EXISTS stream_likes (
    stream_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (stream_id, user_id),
    CONSTRAINT fk_stream_likes_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_stream_likes_user ON stream_likes(user_id);

CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    stream_id UUID NOT NULL,
    user_id UUID NOT NULL,
    message TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_chat_messages_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE,
    CONSTRAINT chk_chat_message_length CHECK (char_length(message) BETWEEN 1 AND 500)
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_stream ON chat_messages(stream_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user ON chat_messages(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS viewer_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    user_id UUID,
    anonymous_id VARCHAR(128),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    watch_seconds INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_viewer_sessions_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE CASCADE,
    CONSTRAINT chk_viewer_identity CHECK (user_id IS NOT NULL OR anonymous_id IS NOT NULL),
    CONSTRAINT chk_viewer_watch_seconds CHECK (watch_seconds >= 0)
);
CREATE INDEX IF NOT EXISTS idx_viewer_sessions_stream ON viewer_sessions(stream_id, joined_at DESC);
CREATE INDEX IF NOT EXISTS idx_viewer_sessions_user ON viewer_sessions(user_id, joined_at DESC);
CREATE INDEX IF NOT EXISTS idx_viewer_sessions_active ON viewer_sessions(stream_id) WHERE left_at IS NULL;

CREATE TABLE IF NOT EXISTS user_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    platform VARCHAR(20) NOT NULL,
    fcm_token TEXT NOT NULL,
    app_version VARCHAR(30),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id)
);
CREATE INDEX IF NOT EXISTS idx_user_devices_user ON user_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_user_devices_active ON user_devices(user_id) WHERE is_active = TRUE;

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(40) NOT NULL,
    actor_user_id UUID,
    channel_id UUID,
    stream_id UUID,
    title VARCHAR(200) NOT NULL,
    body TEXT,
    data JSONB NOT NULL DEFAULT '{}',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_notifications_channel FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE SET NULL,
    CONSTRAINT fk_notifications_stream FOREIGN KEY (stream_id) REFERENCES streams(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, created_at DESC) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_notifications_stream ON notifications(stream_id, created_at DESC);

-- Seed categories
INSERT INTO categories (name, slug, description) VALUES
('Gaming', 'gaming', 'Gaming live streams'),
('Technology', 'technology', 'Technology and software'),
('Programming', 'programming', 'Programming and development'),
('Music', 'music', 'Music and performances'),
('Education', 'education', 'Educational live streams'),
('Sports', 'sports', 'Sports live streams'),
('Entertainment', 'entertainment', 'Entertainment live streams'),
('IRL', 'irl', 'Real-life live streams')
ON CONFLICT (slug) DO NOTHING;

-- Redis is for realtime state, e.g. stream:{id}:viewer_count,
-- stream:{id}:presence, stream:{id}:chat, and stream:{id}:likes.
-- MediaMTX handles RTMP ingest/HLS playback/recording.
-- MinIO stores replay files and images.
-- CDN delivers live HLS and replay content.
