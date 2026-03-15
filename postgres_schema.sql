CREATE TABLE IF NOT EXISTS muted (
    user_id TEXT PRIMARY KEY,
    muted_until TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS monthly_stats (
    month_year CHAR(7) PRIMARY KEY,
    members_joined INT,
    members_left INT
);

CREATE TABLE IF NOT EXISTS streamers (
    discord_id TEXT PRIMARY KEY,
    twitch_id TEXT,
    twitch_name TEXT
);