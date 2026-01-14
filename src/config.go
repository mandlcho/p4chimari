package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	RecentFolders []RecentFolder `json:"recent_folders"`
}

type RecentFolder struct {
	Path      string    `json:"path"`
	LastUsed  time.Time `json:"last_used"`
	UseCount  int       `json:"use_count"`
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".p4chimari.json")
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return empty config if file doesn't exist
		return &Config{RecentFolders: []RecentFolder{}}, nil
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return &Config{RecentFolders: []RecentFolder{}}, nil
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath := getConfigPath()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) AddRecentFolder(path string) {
	// Check if already exists
	for i, folder := range c.RecentFolders {
		if folder.Path == path {
			// Update existing entry
			c.RecentFolders[i].LastUsed = time.Now()
			c.RecentFolders[i].UseCount++
			return
		}
	}

	// Add new entry
	c.RecentFolders = append(c.RecentFolders, RecentFolder{
		Path:     path,
		LastUsed: time.Now(),
		UseCount: 1,
	})

	// Keep only last 10
	if len(c.RecentFolders) > 10 {
		c.RecentFolders = c.RecentFolders[len(c.RecentFolders)-10:]
	}
}

func (c *Config) GetRecentFolders() []RecentFolder {
	// Sort by last used (most recent first)
	sorted := make([]RecentFolder, len(c.RecentFolders))
	copy(sorted, c.RecentFolders)

	// Simple bubble sort by LastUsed
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].LastUsed.Before(sorted[j].LastUsed) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
