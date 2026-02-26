package store

const schema = `
CREATE TABLE IF NOT EXISTS games (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS characters (
    id                TEXT PRIMARY KEY,
    game_id           TEXT NOT NULL UNIQUE REFERENCES games(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    edge              INTEGER NOT NULL CHECK(edge BETWEEN 1 AND 3),
    heart             INTEGER NOT NULL CHECK(heart BETWEEN 1 AND 3),
    iron              INTEGER NOT NULL CHECK(iron BETWEEN 1 AND 3),
    shadow            INTEGER NOT NULL CHECK(shadow BETWEEN 1 AND 3),
    wits              INTEGER NOT NULL CHECK(wits BETWEEN 1 AND 3),
    health            INTEGER NOT NULL DEFAULT 5 CHECK(health BETWEEN 0 AND 5),
    spirit            INTEGER NOT NULL DEFAULT 5 CHECK(spirit BETWEEN 0 AND 5),
    supply            INTEGER NOT NULL DEFAULT 5 CHECK(supply BETWEEN 0 AND 5),
    momentum          INTEGER NOT NULL DEFAULT 2 CHECK(momentum BETWEEN -6 AND 10),
    momentum_max      INTEGER NOT NULL DEFAULT 10,
    momentum_reset    INTEGER NOT NULL DEFAULT 2,
    wounded           INTEGER NOT NULL DEFAULT 0 CHECK(wounded IN (0,1)),
    shaken            INTEGER NOT NULL DEFAULT 0 CHECK(shaken IN (0,1)),
    unprepared        INTEGER NOT NULL DEFAULT 0 CHECK(unprepared IN (0,1)),
    encumbered        INTEGER NOT NULL DEFAULT 0 CHECK(encumbered IN (0,1)),
    maimed            INTEGER NOT NULL DEFAULT 0 CHECK(maimed IN (0,1)),
    corrupted         INTEGER NOT NULL DEFAULT 0 CHECK(corrupted IN (0,1)),
    cursed            INTEGER NOT NULL DEFAULT 0 CHECK(cursed IN (0,1)),
    tormented         INTEGER NOT NULL DEFAULT 0 CHECK(tormented IN (0,1)),
    experience_earned INTEGER NOT NULL DEFAULT 0,
    experience_spent  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS progress_tracks (
    id         TEXT PRIMARY KEY,
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    track_type TEXT NOT NULL CHECK(track_type IN ('vow','journey','combat','bonds')),
    rank       TEXT NOT NULL CHECK(rank IN ('troublesome','dangerous','formidable','extreme','epic')),
    ticks      INTEGER NOT NULL DEFAULT 0 CHECK(ticks BETWEEN 0 AND 40),
    completed  INTEGER NOT NULL DEFAULT 0 CHECK(completed IN (0,1))
);

CREATE TABLE IF NOT EXISTS character_assets (
    id           TEXT PRIMARY KEY,
    character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    asset_id     TEXT NOT NULL,
    name         TEXT NOT NULL,
    ability1     INTEGER NOT NULL DEFAULT 1 CHECK(ability1 IN (0,1)),
    ability2     INTEGER NOT NULL DEFAULT 0 CHECK(ability2 IN (0,1)),
    ability3     INTEGER NOT NULL DEFAULT 0 CHECK(ability3 IN (0,1)),
    health       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS game_log (
    id         TEXT PRIMARY KEY,
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    sequence   INTEGER NOT NULL,
    timestamp  TEXT NOT NULL DEFAULT (datetime('now')),
    entry_type TEXT NOT NULL CHECK(entry_type IN ('move','oracle','narrative','state_change')),
    summary    TEXT NOT NULL,
    details    TEXT NOT NULL DEFAULT '{}',
    tags       TEXT NOT NULL DEFAULT '[]',
    UNIQUE(game_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_game_log_game_seq ON game_log(game_id, sequence);
CREATE INDEX IF NOT EXISTS idx_game_log_type ON game_log(game_id, entry_type);
CREATE INDEX IF NOT EXISTS idx_progress_tracks_game ON progress_tracks(game_id);

CREATE TABLE IF NOT EXISTS combat_state (
    id         TEXT PRIMARY KEY,
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    track_id   TEXT NOT NULL REFERENCES progress_tracks(id) ON DELETE CASCADE,
    foe_name   TEXT NOT NULL,
    initiative INTEGER NOT NULL DEFAULT 0 CHECK(initiative IN (0,1)),
    active     INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0,1))
);
`
