PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at_utc TEXT NOT NULL,
    updated_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at_utc TEXT NOT NULL,
    updated_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS saved_charts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    chart_type TEXT NOT NULL DEFAULT 'natal',
    project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
    client_id TEXT REFERENCES clients(id) ON DELETE SET NULL,
    local_date TEXT NOT NULL,
    local_time TEXT NOT NULL,
    utc_offset TEXT NOT NULL,
    latitude_degrees TEXT NOT NULL,
    longitude_degrees TEXT NOT NULL,
    location_name TEXT NOT NULL DEFAULT '',
    source_system TEXT NOT NULL DEFAULT 'astro-go',
    created_at_utc TEXT NOT NULL DEFAULT '',
    updated_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chart_versions (
    id TEXT PRIMARY KEY,
    chart_id TEXT NOT NULL REFERENCES saved_charts(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    payload_json TEXT NOT NULL,
    created_at_utc TEXT NOT NULL,
    UNIQUE(chart_id, version_number)
);

CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '',
    created_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chart_tags (
    chart_id TEXT NOT NULL REFERENCES saved_charts(id) ON DELETE CASCADE,
    tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY(chart_id, tag_id)
);

CREATE TABLE IF NOT EXISTS chart_notes (
    id TEXT PRIMARY KEY,
    chart_id TEXT NOT NULL REFERENCES saved_charts(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    created_at_utc TEXT NOT NULL,
    updated_at_utc TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS interpretation_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    module TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at_utc TEXT NOT NULL,
    updated_at_utc TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_saved_charts_updated_at ON saved_charts(updated_at_utc DESC);
CREATE INDEX IF NOT EXISTS idx_saved_charts_project ON saved_charts(project_id);
CREATE INDEX IF NOT EXISTS idx_saved_charts_client ON saved_charts(client_id);
CREATE INDEX IF NOT EXISTS idx_chart_versions_chart ON chart_versions(chart_id);
CREATE INDEX IF NOT EXISTS idx_chart_notes_chart ON chart_notes(chart_id);
