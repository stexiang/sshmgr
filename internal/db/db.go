package db

import (
	"database/sql"
	"time"
)

func Init(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return err
	}

	schema := `
CREATE TABLE IF NOT EXISTS hosts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  user TEXT NOT NULL,
  host TEXT NOT NULL,
  port INTEGER NOT NULL DEFAULT 22,
  note TEXT DEFAULT '',
  tags TEXT DEFAULT '',
  last_ip TEXT DEFAULT '',
  last_checked_at TEXT DEFAULT '',
  has_secret INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS conn_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  host_id INTEGER NOT NULL,
  start_at TEXT NOT NULL,
  end_at TEXT NOT NULL,
  duration_ms INTEGER NOT NULL,
  resolved_ip TEXT DEFAULT '',
  exit_code INTEGER NOT NULL,
  local_user TEXT DEFAULT '',
  FOREIGN KEY(host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_conn_log_host_time ON conn_log(host_id, start_at);
`
	_, err := db.Exec(schema)
	return err
}

func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
