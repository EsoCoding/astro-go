package storage

type Settings struct {
	DateFormat      string
	TimeFormat      string
	DefaultLocation string
	DefaultLat      string
	DefaultLng      string
	NodePreference  string
	PoFPreference   string
}

func (s *ChartStore) GetSettings() (Settings, error) {
	settings := Settings{
		DateFormat:      "YYYY-MM-DD",
		TimeFormat:      "24h",
		DefaultLocation: "",
		DefaultLat:      "",
		DefaultLng:      "",
		NodePreference:  "True",
		PoFPreference:   "Day",
	}

	if s == nil || s.db == nil {
		return settings, nil
	}

	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		// Table might not exist yet, or other error. Just return defaults.
		return settings, nil
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err == nil {
			switch key {
			case "DateFormat":
				settings.DateFormat = value
			case "TimeFormat":
				settings.TimeFormat = value
			case "DefaultLocation":
				settings.DefaultLocation = value
			case "DefaultLat":
				settings.DefaultLat = value
			case "DefaultLng":
				settings.DefaultLng = value
			case "NodePreference":
				settings.NodePreference = value
			case "PoFPreference":
				settings.PoFPreference = value
			}
		}
	}

	return settings, nil
}

func (s *ChartStore) SaveSettings(settings Settings) error {
	if s == nil || s.db == nil {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO settings (key, value, updated_at_utc) VALUES (?, ?, datetime('now')) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at_utc=excluded.updated_at_utc`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	kv := map[string]string{
		"DateFormat":      settings.DateFormat,
		"TimeFormat":      settings.TimeFormat,
		"DefaultLocation": settings.DefaultLocation,
		"DefaultLat":      settings.DefaultLat,
		"DefaultLng":      settings.DefaultLng,
		"NodePreference":  settings.NodePreference,
		"PoFPreference":   settings.PoFPreference,
	}

	for k, v := range kv {
		if _, err := stmt.Exec(k, v); err != nil {
			return err
		}
	}

	return tx.Commit()
}
