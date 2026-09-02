package config

import (
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

type ArchscanConfig struct {
	Exclude    []string   `yaml:"exclude"`
	Thresholds Thresholds `yaml:"thresholds"`
	Rules      Rules      `yaml:"rules"`
}

type Thresholds struct {
	GodFunctionLines  int `yaml:"god_function_lines"`
	DuplicateMinLines int `yaml:"duplicate_min_lines"`
	MaxFileViolations int `yaml:"max_file_violations"`
}

type Rules struct {
	CheckBoundaries   bool `yaml:"check_boundaries"`
	CheckNaming       bool `yaml:"check_naming"`
	CheckDuplication  bool `yaml:"check_duplication"`
	CheckAntipatterns bool `yaml:"check_antipatterns"`
	CheckDeadCode     bool `yaml:"check_dead_code"`
}

func DefaultConfig() *ArchscanConfig {
	return &ArchscanConfig{
		Exclude: []string{},
		Thresholds: Thresholds{
			GodFunctionLines:  80,
			DuplicateMinLines: 8,
			MaxFileViolations: 20,
		},
		Rules: Rules{
			CheckBoundaries:   true,
			CheckNaming:       true,
			CheckDuplication:  true,
			CheckAntipatterns: true,
			CheckDeadCode:     true,
		},
	}
}

func Load(repoPath string) *ArchscanConfig {
	cfg := DefaultConfig()
	data, err := os.ReadFile(filepath.Join(repoPath, "archscan.yaml"))
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, cfg)
	return cfg
}
