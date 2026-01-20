package reassoc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Entry struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type Table struct {
	mu   sync.Mutex
	Path string
	Map  map[string]Entry // fingerprint -> entry
}

func Load() (*Table, error) {
	cfg, _ := os.UserConfigDir()
	path := filepath.Join(cfg, "sshmgr", "reassoc.json")

	t := &Table{
		Path: path,
		Map:  map[string]Entry{},
	}

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &t.Map)
	}

	return t, nil
}

func (t *Table) Save() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	_ = os.MkdirAll(filepath.Dir(t.Path), 0o755)

	data, _ := json.MarshalIndent(t.Map, "", "  ")
	return os.WriteFile(t.Path, data, 0o644)
}

func (t *Table) Lookup(fp string) (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.Map[fp]
	return e, ok
}

func (t *Table) Update(fp, name, ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Map[fp] = Entry{Name: name, IP: ip}
}
