package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"example.com/shop/stock"
)

// load reads a stock file written by save.
func load(path string) (*stock.Stock, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s not found; run \"shop init\" first", path)
	}
	if err != nil {
		return nil, err
	}
	var counts map[string]int
	if err := json.Unmarshal(data, &counts); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	st := stock.NewStock()
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		st.Add(name, counts[name])
	}
	return st, nil
}

// save writes st to path as a JSON object of item counts.
func save(path string, st *stock.Stock) error {
	data, err := json.MarshalIndent(st.Items(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
