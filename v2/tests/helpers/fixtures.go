package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FixtureLoader provides utilities for loading test fixture files
type FixtureLoader struct {
	fixturesDir string
}

// NewFixtureLoader creates a new fixture loader pointing to the test fixtures directory
func NewFixtureLoader() *FixtureLoader {
	// Try to find the fixtures directory relative to the test file
	// Assumes tests are in /v2/tests/helpers/
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Try multiple paths to find fixtures
	possiblePaths := []string{
		filepath.Join(cwd, "fixtures"),
		filepath.Join(cwd, "tests", "fixtures"),
		filepath.Join(cwd, "..", "fixtures"),
		filepath.Join(cwd, "..", "..", "v2", "tests", "fixtures"),
	}

	for _, path := range possiblePaths {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			return &FixtureLoader{fixturesDir: path}
		}
	}

	// Default to current directory if nothing found
	return &FixtureLoader{fixturesDir: "."}
}

// WithFixturesDir creates a FixtureLoader with a custom fixtures directory
func NewFixtureLoaderWithDir(dir string) *FixtureLoader {
	return &FixtureLoader{fixturesDir: dir}
}

// LoadJSONFixture loads and parses a JSON fixture file
func (fl *FixtureLoader) LoadJSONFixture(filename string, v interface{}) error {
	path := filepath.Join(fl.fixturesDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read fixture file %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal JSON fixture %s: %w", filename, err)
	}

	return nil
}

// LoadTextFixture loads a text fixture file
func (fl *FixtureLoader) LoadTextFixture(filename string) (string, error) {
	path := filepath.Join(fl.fixturesDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read fixture file %s: %w", filename, err)
	}
	return string(data), nil
}

// LoadBinaryFixture loads a binary fixture file
func (fl *FixtureLoader) LoadBinaryFixture(filename string) ([]byte, error) {
	path := filepath.Join(fl.fixturesDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fixture file %s: %w", filename, err)
	}
	return data, nil
}

// SaveFixture saves data to a fixture file (useful for creating fixtures)
func (fl *FixtureLoader) SaveJSONFixture(filename string, v interface{}) error {
	path := filepath.Join(fl.fixturesDir, filename)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create fixture directory: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON fixture: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write fixture file %s: %w", filename, err)
	}

	return nil
}

// FixturesDir returns the fixtures directory path
func (fl *FixtureLoader) FixturesDir() string {
	return fl.fixturesDir
}

// Exists checks if a fixture file exists
func (fl *FixtureLoader) Exists(filename string) bool {
	path := filepath.Join(fl.fixturesDir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// ODSFixture provides fixture utilities for ODS files
type ODSFixture struct {
	loader *FixtureLoader
}

// NewODSFixture creates an ODS fixture helper
func NewODSFixture() *ODSFixture {
	return &ODSFixture{loader: NewFixtureLoader()}
}

// LoadODSFile loads an ODS file from fixtures
func (of *ODSFixture) LoadODSFile(filename string) ([]byte, error) {
	return of.loader.LoadBinaryFixture(filepath.Join("ods", filename))
}

// ListODSFixtures lists all ODS fixture files
func (of *ODSFixture) ListODSFixtures() ([]string, error) {
	odsDir := filepath.Join(of.loader.FixturesDir(), "ods")
	entries, err := os.ReadDir(odsDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// HTMLFixture provides fixture utilities for HTML files
type HTMLFixture struct {
	loader *FixtureLoader
}

// NewHTMLFixture creates an HTML fixture helper
func NewHTMLFixture() *HTMLFixture {
	return &HTMLFixture{loader: NewFixtureLoader()}
}

// LoadHTMLFile loads an HTML file from fixtures
func (hf *HTMLFixture) LoadHTMLFile(filename string) (string, error) {
	return hf.loader.LoadTextFixture(filepath.Join("html", filename))
}

// ListHTMLFixtures lists all HTML fixture files
func (hf *HTMLFixture) ListHTMLFixtures() ([]string, error) {
	htmlDir := filepath.Join(hf.loader.FixturesDir(), "html")
	entries, err := os.ReadDir(htmlDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// EntityFixture provides fixture utilities for entity JSON files
type EntityFixture struct {
	loader *FixtureLoader
}

// NewEntityFixture creates an entity fixture helper
func NewEntityFixture() *EntityFixture {
	return &EntityFixture{loader: NewFixtureLoader()}
}

// LoadEntityFixture loads an entity fixture JSON file
func (ef *EntityFixture) LoadEntityFixture(filename string, v interface{}) error {
	return ef.loader.LoadJSONFixture(filepath.Join("entities", filename), v)
}

// SaveEntityFixture saves an entity as a JSON fixture
func (ef *EntityFixture) SaveEntityFixture(filename string, v interface{}) error {
	return ef.loader.SaveJSONFixture(filepath.Join("entities", filename), v)
}

// ListEntityFixtures lists all entity fixture files
func (ef *EntityFixture) ListEntityFixtures() ([]string, error) {
	entitiesDir := filepath.Join(ef.loader.FixturesDir(), "entities")
	entries, err := os.ReadDir(entitiesDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// DataFixture provides fixture utilities for generic data files
type DataFixture struct {
	loader *FixtureLoader
}

// NewDataFixture creates a data fixture helper
func NewDataFixture() *DataFixture {
	return &DataFixture{loader: NewFixtureLoader()}
}

// LoadData loads a data fixture file
func (df *DataFixture) LoadData(filename string) ([]byte, error) {
	return df.loader.LoadBinaryFixture(filepath.Join("data", filename))
}

// LoadDataText loads a text data fixture file
func (df *DataFixture) LoadDataText(filename string) (string, error) {
	return df.loader.LoadTextFixture(filepath.Join("data", filename))
}

// SaveData saves data to a fixture file
func (df *DataFixture) SaveData(filename string, data []byte) error {
	path := filepath.Join(df.loader.FixturesDir(), "data", filename)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data fixture directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ListDataFixtures lists all data fixture files
func (df *DataFixture) ListDataFixtures() ([]string, error) {
	dataDir := filepath.Join(df.loader.FixturesDir(), "data")
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
