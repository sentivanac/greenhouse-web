package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"greenhouse/internal/model"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	s := &SQLite{db: db}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SQLite) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS measurements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ts INTEGER NOT NULL,
		temperature REAL,
		humidity REAL,
		pressure REAL,
		light INTEGER,
		soil INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_measurements_ts
	ON measurements(ts);
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLite) Insert(m model.Measurement) error {
	query := `
	INSERT INTO measurements (ts, temperature, humidity, pressure, light, soil)
	VALUES (?, ?, ?, ?, ?, ?);
	`

	_, err := s.db.Exec(
		query,
		m.Ts,
		m.T,
		m.RH,
		m.P,
		m.Light,
		m.Soil,
	)

	return err
}

func (s *SQLite) Count() (int, error) {
	row := s.db.QueryRow("SELECT COUNT(*) FROM measurements")
	var c int
	err := row.Scan(&c)
	return c, err
}

func (s *SQLite) GetLatest() (model.Measurement, error) {
	query := `
	SELECT ts, temperature, humidity, pressure, light, soil
	FROM measurements
	ORDER BY ts DESC
	LIMIT 1;
	`

	row := s.db.QueryRow(query)

	var m model.Measurement
	err := row.Scan(
		&m.Ts,
		&m.T,
		&m.RH,
		&m.P,
		&m.Light,
		&m.Soil,
	)

	return m, err
}

func (s *SQLite) GetRecent(limit int) ([]model.Measurement, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
	SELECT ts, temperature, humidity, pressure, light, soil
	FROM measurements
	ORDER BY ts DESC
	LIMIT ?;
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.Measurement
	for rows.Next() {
		var m model.Measurement
		if err := rows.Scan(
			&m.Ts,
			&m.T,
			&m.RH,
			&m.P,
			&m.Light,
			&m.Soil,
		); err != nil {
			return nil, err
		}
		res = append(res, m)
	}

	return res, rows.Err()
}

func (s *SQLite) GetRange(from, to int64) ([]model.Measurement, error) {
	query := `
	SELECT ts, temperature, humidity, pressure, light, soil
	FROM measurements
	WHERE ts >= ? AND ts <= ?
	ORDER BY ts;
	`

	rows, err := s.db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.Measurement
	for rows.Next() {
		var m model.Measurement
		if err := rows.Scan(&m.Ts, &m.T, &m.RH, &m.P, &m.Light, &m.Soil); err != nil {
			return nil, err
		}
		res = append(res, m)
	}

	return res, rows.Err()
}

func (s *SQLite) GetRangeDownsampled(from, to int64) ([]model.Measurement, int64, error) {
	step := pickStep(from, to)

	query := `
	SELECT
		(ts / ?) * ?            AS bucket,
		AVG(temperature)        AS temperature,
		AVG(humidity)           AS humidity,
		AVG(pressure)           AS pressure,
		AVG(light)              AS light,
		AVG(soil)               AS soil,
	FROM measurements
	WHERE ts BETWEEN ? AND ?
	GROUP BY bucket
	ORDER BY bucket;
	`

	rows, err := s.db.Query(query, step, step, from, to)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var res []model.Measurement
	for rows.Next() {
		var m model.Measurement
		if err := rows.Scan(&m.Ts, &m.T, &m.RH, &m.P, &m.Light, &m.Soil); err != nil {
			return nil, 0, err
		}
		res = append(res, m)
	}

	return res, step, rows.Err()
}

func (s *SQLite) GetRangeEnvelope(from, to int64) ([]model.Envelope, int64, error) {
	step := pickStep(from, to)

	query := `
	SELECT
		(ts / ?) * ? AS bucket,

		AVG(temperature),
		MIN(temperature),
		MAX(temperature),

		AVG(humidity),
		MIN(humidity),
		MAX(humidity),

		AVG(light),
		MIN(light),
		MAX(light)

		AVG(soil),
		MIN(soil),
		MAX(soil)

	FROM measurements
	WHERE ts BETWEEN ? AND ?
	GROUP BY bucket
	ORDER BY bucket;
	`

	rows, err := s.db.Query(query, step, step, from, to)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var res []model.Envelope
	for rows.Next() {
		var e model.Envelope
		if err := rows.Scan(
			&e.Ts,
			&e.TempAvg, &e.TempMin, &e.TempMax,
			&e.HumAvg, &e.HumMin, &e.HumMax,
			&e.LightAvg, &e.LightMin, &e.LightMax,
			&e.SoilAvg, &e.SoilMin, &e.SoilMax,
		); err != nil {
			return nil, 0, err
		}
		res = append(res, e)
	}

	return res, step, rows.Err()
}
