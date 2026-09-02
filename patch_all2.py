import re

def patch(path, pattern, repl):
    with open(path, "r") as f: text = f.read()
    new_text, n = re.subn(pattern, repl, text, flags=re.MULTILINE)
    if n > 0:
        with open(path, "w") as f: f.write(new_text)
        print(f"Patched {path}")
    else:
        print(f"Failed to patch {path}")

# analyzer.go
a_path = "internal/analyzer/analyzer.go"
patch(a_path, r"func Analyze\(repoPath string, verbose bool\) \(\*Result, error\) {", r"func Analyze(repoPath string, verbose bool, cfg *config.ArchscanConfig) (*Result, error) {")
patch(a_path, r"files, err := collectFiles\(repoPath, lang\)", r"files, err := collectFiles(repoPath, lang, cfg.Exclude)")
patch(a_path, r"violations, err := d\.Detect\(repoPath, files, lang, verbose\)", r"violations, err := d.Detect(repoPath, files, lang, verbose, cfg)")
patch(a_path, r"\"strings\"", "\"strings\"\n\t\"archscan/internal/config\"")

# detector.go
d_path = "internal/analyzer/detector.go"
patch(d_path, r"Detect\(root string, files \[\]string, lang string, verbose bool\) \(\[\]Violation, error\)", r"Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig) ([]Violation, error)")
patch(d_path, r"package analyzer", "package analyzer\n\nimport \"archscan/internal/config\"")

# others
for f in ["antipattern.go", "boundary.go", "duplication.go", "naming.go"]:
    p = "internal/analyzer/" + f
    patch(p, r"Detect\(root string, files \[\]string, lang string, verbose bool\) \(\[\]Violation, error\)", r"Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig) ([]Violation, error)")
    patch(p, r"\"strings\"", "\"strings\"\n\t\"archscan/internal/config\"")
