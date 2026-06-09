package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"astro-go/internal/astro"

	"fyne.io/fyne/v2"
	_ "modernc.org/sqlite"
)

const (
	savedChartsKey        = "saved_charts_v1"
	sqliteMigrationKey    = "saved_charts_sqlite_imported_v1"
	defaultDatabaseFolder = "astro-go"
	defaultDatabaseFile   = "charts.sqlite"
)

type SavedChart struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ChartType         string `json:"chart_type"`
	HouseSystem       string `json:"house_system"`
	BaseChartID       string `json:"base_chart_id"`
	ComparisonChartID string `json:"comparison_chart_id"`
	ReferenceDate     string `json:"reference_date"`
	ReferenceTime     string `json:"reference_time"`
	ReferenceUTC      string `json:"reference_datetime_utc"`
	LocalDate         string `json:"local_date"`
	LocalTime         string `json:"local_time"`
	UTCOffset         string `json:"utc_offset"`
	LocationName      string `json:"location_name"`
	LatitudeDegrees   string `json:"latitude_degrees"`
	LongitudeDegrees  string `json:"longitude_degrees"`
	UpdatedAtUTC      string `json:"updated_at_utc"`
}

func SavedChartFromBirthData(data astro.BirthData, localDate, localTime, utcOffset, locationName, latitude, longitude string) SavedChart {
	now := time.Now().UTC()
	return SavedChart{
		ID:               chartID(data.Name, now),
		Name:             data.Name,
		ChartType:        string(astro.ChartTypeNatal),
		HouseSystem:      string(data.HouseSystem),
		LocalDate:        localDate,
		LocalTime:        localTime,
		UTCOffset:        utcOffset,
		LocationName:     locationName,
		LatitudeDegrees:  latitude,
		LongitudeDegrees: longitude,
		UpdatedAtUTC:     now.Format(time.RFC3339),
	}
}

type ChartStore struct {
	db   *sql.DB
	path string
}

func DefaultDatabasePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, defaultDatabaseFolder, defaultDatabaseFile), nil
}

func NewChartStore(preferences fyne.Preferences) (*ChartStore, error) {
	path, err := DefaultDatabasePath()
	if err != nil {
		return nil, err
	}
	store, err := NewSQLiteChartStore(path)
	if err != nil {
		return nil, err
	}
	if err := store.ImportPreferenceCharts(preferences); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

func NewSQLiteChartStore(path string) (*ChartStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &ChartStore{db: db, path: path}
	if err := store.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *ChartStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *ChartStore) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *ChartStore) List() ([]SavedChart, error) {
	rows, err := s.db.Query(`
		SELECT id, name, chart_type, house_system, base_chart_id, comparison_chart_id, reference_date, reference_time, reference_datetime_utc,
		       local_date, local_time, utc_offset, location_name, latitude_degrees, longitude_degrees, updated_at_utc
		FROM saved_charts
		ORDER BY updated_at_utc DESC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []SavedChart
	for rows.Next() {
		var chart SavedChart
		var baseChartID sql.NullString
		var comparisonChartID sql.NullString
		if err := rows.Scan(
			&chart.ID,
			&chart.Name,
			&chart.ChartType,
			&chart.HouseSystem,
			&baseChartID,
			&comparisonChartID,
			&chart.ReferenceDate,
			&chart.ReferenceTime,
			&chart.ReferenceUTC,
			&chart.LocalDate,
			&chart.LocalTime,
			&chart.UTCOffset,
			&chart.LocationName,
			&chart.LatitudeDegrees,
			&chart.LongitudeDegrees,
			&chart.UpdatedAtUTC,
		); err != nil {
			return nil, err
		}
		chart.BaseChartID = baseChartID.String
		chart.ComparisonChartID = comparisonChartID.String
		charts = append(charts, chart)
	}
	return charts, rows.Err()
}

func (s *ChartStore) Save(chart *SavedChart) error {
	return s.save(chart, true)
}

func (s *ChartStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM saved_charts WHERE id = ?`, id)
	return err
}

func (s *ChartStore) ImportPreferenceCharts(preferences fyne.Preferences) error {
	if preferences == nil {
		return nil
	}
	if preferences.StringWithFallback(sqliteMigrationKey, "") == "done" {
		return nil
	}
	raw := preferences.StringWithFallback(savedChartsKey, "[]")
	var charts []SavedChart
	if err := json.Unmarshal([]byte(raw), &charts); err != nil {
		return err
	}
	for _, chart := range charts {
		if err := s.save(&chart, false); err != nil {
			return err
		}
	}
	preferences.SetString(sqliteMigrationKey, "done")
	return nil
}

func (s *ChartStore) initSchema() error {
	_, err := s.db.Exec(`
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
			house_system TEXT NOT NULL DEFAULT 'W',
			base_chart_id TEXT REFERENCES saved_charts(id) ON DELETE SET NULL,
			comparison_chart_id TEXT REFERENCES saved_charts(id) ON DELETE SET NULL,
			reference_date TEXT NOT NULL DEFAULT '',
			reference_time TEXT NOT NULL DEFAULT '',
			reference_datetime_utc TEXT NOT NULL DEFAULT '',
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

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at_utc TEXT NOT NULL
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
	`)
	if err != nil {
		return err
	}
	for _, column := range []struct {
		name       string
		definition string
	}{
		{"chart_type", "TEXT NOT NULL DEFAULT 'natal'"},
		{"house_system", "TEXT NOT NULL DEFAULT 'W'"},
		{"base_chart_id", "TEXT REFERENCES saved_charts(id) ON DELETE SET NULL"},
		{"comparison_chart_id", "TEXT REFERENCES saved_charts(id) ON DELETE SET NULL"},
		{"reference_date", "TEXT NOT NULL DEFAULT ''"},
		{"reference_time", "TEXT NOT NULL DEFAULT ''"},
		{"reference_datetime_utc", "TEXT NOT NULL DEFAULT ''"},
		{"project_id", "TEXT REFERENCES projects(id) ON DELETE SET NULL"},
		{"client_id", "TEXT REFERENCES clients(id) ON DELETE SET NULL"},
		{"location_name", "TEXT NOT NULL DEFAULT ''"},
		{"source_system", "TEXT NOT NULL DEFAULT 'astro-go'"},
		{"created_at_utc", "TEXT NOT NULL DEFAULT ''"},
	} {
		if err := s.addSavedChartColumnIfMissing(column.name, column.definition); err != nil {
			return err
		}
	}
	return s.initIndexes()
}

func (s *ChartStore) initIndexes() error {
	_, err := s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_saved_charts_updated_at ON saved_charts(updated_at_utc DESC);
		CREATE INDEX IF NOT EXISTS idx_saved_charts_type ON saved_charts(chart_type);
		CREATE INDEX IF NOT EXISTS idx_saved_charts_house_system ON saved_charts(house_system);
		CREATE INDEX IF NOT EXISTS idx_saved_charts_base ON saved_charts(base_chart_id);
		CREATE INDEX IF NOT EXISTS idx_saved_charts_comparison ON saved_charts(comparison_chart_id);
		CREATE INDEX IF NOT EXISTS idx_saved_charts_project ON saved_charts(project_id);
		CREATE INDEX IF NOT EXISTS idx_saved_charts_client ON saved_charts(client_id);
		CREATE INDEX IF NOT EXISTS idx_chart_versions_chart ON chart_versions(chart_id);
		CREATE INDEX IF NOT EXISTS idx_chart_notes_chart ON chart_notes(chart_id);
	`)
	return err
}

func (s *ChartStore) addSavedChartColumnIfMissing(name, definition string) error {
	rows, err := s.db.Query(`PRAGMA table_info(saved_charts)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var columnName, columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &columnName, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if columnName == name {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.db.Exec(fmt.Sprintf(`ALTER TABLE saved_charts ADD COLUMN %s %s`, name, definition))
	return err
}

func (s *ChartStore) save(chart *SavedChart, touchUpdatedAt bool) error {
	now := time.Now().UTC()
	if chart.ID == "" {
		chart.ID = chartID(chart.Name, now)
	}
	if chart.ChartType == "" {
		chart.ChartType = string(astro.ChartTypeNatal)
	}
	if chart.HouseSystem == "" {
		chart.HouseSystem = string(astro.DefaultHouseSystem())
	}
	if touchUpdatedAt || chart.UpdatedAtUTC == "" {
		chart.UpdatedAtUTC = now.Format(time.RFC3339Nano)
	}
	_, err := s.db.Exec(`
		INSERT INTO saved_charts (
			id, name, chart_type, house_system, base_chart_id, comparison_chart_id, reference_date, reference_time, reference_datetime_utc,
			local_date, local_time, utc_offset, location_name, latitude_degrees, longitude_degrees, updated_at_utc
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			chart_type = excluded.chart_type,
			house_system = excluded.house_system,
			base_chart_id = excluded.base_chart_id,
			comparison_chart_id = excluded.comparison_chart_id,
			reference_date = excluded.reference_date,
			reference_time = excluded.reference_time,
			reference_datetime_utc = excluded.reference_datetime_utc,
			local_date = excluded.local_date,
			local_time = excluded.local_time,
			utc_offset = excluded.utc_offset,
			location_name = excluded.location_name,
			latitude_degrees = excluded.latitude_degrees,
			longitude_degrees = excluded.longitude_degrees,
			updated_at_utc = excluded.updated_at_utc
	`, chart.ID, chart.Name, chart.ChartType, chart.HouseSystem, nullableString(chart.BaseChartID), nullableString(chart.ComparisonChartID), chart.ReferenceDate, chart.ReferenceTime, chart.ReferenceUTC, chart.LocalDate, chart.LocalTime, chart.UTCOffset, chart.LocationName, chart.LatitudeDegrees, chart.LongitudeDegrees, chart.UpdatedAtUTC)
	return err
}

func chartID(name string, t time.Time) string {
	if name == "" {
		name = "chart"
	}
	return fmt.Sprintf("%s-%d", name, t.UnixNano())
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
