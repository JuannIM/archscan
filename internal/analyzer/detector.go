package analyzer

// Detector is the interface all violation detectors must implement.
type Detector interface {
	Detect(root string, files []string, lang string, verbose bool) ([]Violation, error)
}
