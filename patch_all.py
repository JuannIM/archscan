import os
def patch_file(path, old, new):
    if not os.path.exists(path): return
    with open(path, "r") as f: text = f.read()
    if old in text:
        text = text.replace(old, new)
        with open(path, "w") as f: f.write(text)

# analyzer.go
a_path = "internal/analyzer/analyzer.go"
patch_file(a_path, "func Analyze(repoPath string, verbose bool) (*Result, error) {", "func Analyze(repoPath string, verbose bool, cfg *config.ArchscanConfig) (*Result, error) {")
patch_file(a_path, "files, err := collectFiles(repoPath, lang)", "files, err := collectFiles(repoPath, lang, cfg.Exclude)")
patch_file(a_path, "violations, err := d.Detect(repoPath, files, lang, verbose)", "violations, err := d.Detect(repoPath, files, lang, verbose, cfg)")

# detector.go
d_path = "internal/analyzer/detector.go"
patch_file(d_path, "Detect(root string, files []string, lang string, verbose bool) ([]Violation, error)", "Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig) ([]Violation, error)")

# others
for f in ["antipattern.go", "boundary.go", "duplication.go", "naming.go"]:
    p = "internal/analyzer/" + f
    patch_file(p, "Detect(root string, files []string, lang string, verbose bool)", "Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig)")

