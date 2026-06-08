package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestSQLiteChartStoreSaveListAndDelete(t *testing.T) {
	store, err := NewSQLiteChartStore(filepath.Join(t.TempDir(), "charts.sqlite"))
	if err != nil {
		t.Fatalf("NewSQLiteChartStore() error = %v", err)
	}
	defer store.Close()

	first := SavedChart{
		ID:               "first",
		Name:             "First Chart",
		ChartType:        "natal",
		HouseSystem:      "W",
		LocationName:     "Amsterdam, Netherlands",
		LocalDate:        "1990-01-01",
		LocalTime:        "12:00",
		UTCOffset:        "0",
		LatitudeDegrees:  "52.3676",
		LongitudeDegrees: "4.9041",
		UpdatedAtUTC:     time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	second := SavedChart{
		ID:               "second",
		Name:             "Second Chart",
		ChartType:        "transit",
		HouseSystem:      "P",
		BaseChartID:      "first",
		ReferenceDate:    "2026-01-02",
		ReferenceTime:    "10:00",
		ReferenceUTC:     time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
		LocationName:     "Rotterdam, Netherlands",
		LocalDate:        "1991-02-03",
		LocalTime:        "14:30",
		UTCOffset:        "1",
		LatitudeDegrees:  "51.9244",
		LongitudeDegrees: "4.4777",
		UpdatedAtUTC:     time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}

	if err := store.save(first, false); err != nil {
		t.Fatalf("save(first) error = %v", err)
	}
	if err := store.save(second, false); err != nil {
		t.Fatalf("save(second) error = %v", err)
	}

	charts, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(charts) != 2 {
		t.Fatalf("List() len = %d, want 2", len(charts))
	}
	if charts[0].ID != "second" || charts[1].ID != "first" {
		t.Fatalf("List() order = %q, %q; want second, first", charts[0].ID, charts[1].ID)
	}
	if charts[0].ChartType != "transit" || charts[0].HouseSystem != "P" || charts[0].BaseChartID != "first" || charts[0].ReferenceDate != "2026-01-02" {
		t.Fatalf("transit fields not persisted: %+v", charts[0])
	}

	first.Name = "First Chart Updated"
	if err := store.Save(first); err != nil {
		t.Fatalf("Save(update) error = %v", err)
	}
	charts, err = store.List()
	if err != nil {
		t.Fatalf("List() after update error = %v", err)
	}
	if len(charts) != 2 {
		t.Fatalf("List() after update len = %d, want 2", len(charts))
	}
	if charts[0].ID != "first" || charts[0].Name != "First Chart Updated" {
		t.Fatalf("updated chart = %+v, want first chart first with updated name", charts[0])
	}

	if err := store.Delete("first"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	charts, err = store.List()
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(charts) != 1 || charts[0].ID != "second" {
		t.Fatalf("remaining charts = %+v, want only second", charts)
	}
}

func TestSQLiteChartStoreInitializesApplicationTables(t *testing.T) {
	store, err := NewSQLiteChartStore(filepath.Join(t.TempDir(), "charts.sqlite"))
	if err != nil {
		t.Fatalf("NewSQLiteChartStore() error = %v", err)
	}
	defer store.Close()

	for _, table := range []string{
		"projects",
		"clients",
		"saved_charts",
		"chart_versions",
		"tags",
		"chart_tags",
		"chart_notes",
		"interpretation_templates",
	} {
		var name string
		err := store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %q missing: %v", table, err)
		}
	}
}

func TestSQLiteChartStoreMigratesOldSavedChartsSchemaBeforeIndexing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "charts.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE saved_charts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			local_date TEXT NOT NULL,
			local_time TEXT NOT NULL,
			utc_offset TEXT NOT NULL,
			latitude_degrees TEXT NOT NULL,
			longitude_degrees TEXT NOT NULL,
			updated_at_utc TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create old schema error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close old db error = %v", err)
	}

	store, err := NewSQLiteChartStore(path)
	if err != nil {
		t.Fatalf("NewSQLiteChartStore() with old schema error = %v", err)
	}
	defer store.Close()

	for _, column := range []string{"project_id", "client_id", "chart_type", "house_system", "base_chart_id", "comparison_chart_id", "reference_date", "reference_time", "reference_datetime_utc", "location_name", "source_system", "created_at_utc"} {
		if !savedChartsColumnExists(t, store, column) {
			t.Fatalf("expected migrated column %q", column)
		}
	}
}

func savedChartsColumnExists(t *testing.T, store *ChartStore, column string) bool {
	t.Helper()

	rows, err := store.db.Query(`PRAGMA table_info(saved_charts)`)
	if err != nil {
		t.Fatalf("PRAGMA table_info(saved_charts) error = %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var columnName, columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &columnName, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan table info error = %v", err)
		}
		if columnName == column {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error = %v", err)
	}
	return false
}
