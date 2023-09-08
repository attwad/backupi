package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Path []string    `yaml:"path"`
	Dest Destination `yaml:"dest"`
}

type Destination struct {
	LocalDir *LocalDest `yaml:"localDir,omitempty"`
	GCS      *GCS       `yaml:"gcs,omitempty"`
}

type LocalDest struct {
	// Path is the absolute path to a directory in which to store the backup.tar file.
	Path string `yaml:"path"`
}

type GCS struct {
	// Bucket is the bucket in which to put the backup.tar file.
	Bucket string `yaml:"bucket"`
}

func Read(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", configPath, err)
	}
	c := &Config{}
	fmt.Printf("config:\n%s\n", string(data))
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshaling config yaml: %v", err)
	}
	return c, nil
}
