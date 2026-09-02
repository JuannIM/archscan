package analyzer

import "archscan/internal/config"

// Detector is the interface all violation detectors must implement.
type Detector interface {
	Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig) ([]Violation, error)
}
