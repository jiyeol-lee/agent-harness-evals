package stock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Load reads a stock file written by Save.
func Load(path string) (*Stock, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s not found; run \"tally init\" first", path)
	}
	if err != nil {
		return nil, err
	}
	s := New()
	if err := json.Unmarshal(data, &s.qty); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// Save writes s to path as a JSON object of item counts.
func (s *Stock) Save(path string) error {
	data, err := json.MarshalIndent(s.qty, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
