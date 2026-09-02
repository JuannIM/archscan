python3 -c '
import re
path = "internal/analyzer/analyzer.go"
with open(path, "r") as f: text = f.read()

# Add isExcluded and use it in collectFiles
old_collect = """		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}"""
new_collect = """		rel, _ := filepath.Rel(root, path)
		if info.IsDir() {
			if shouldSkipDir(info.Name()) || isExcluded(rel, exclude) {
				return filepath.SkipDir
			}
			return nil
		}
		if isExcluded(rel, exclude) {
			return nil
		}"""
text = text.replace(old_collect, new_collect)

is_ex_func = """
func isExcluded(relPath string, exclude []string) bool {
	slashPath := filepath.ToSlash(relPath)
	for _, ex := range exclude {
		if strings.HasPrefix(ex, "**/") && strings.HasSuffix(ex, "/**") {
			match := strings.TrimSuffix(strings.TrimPrefix(ex, "**/"), "/**")
			if strings.Contains(slashPath, "/"+match+"/") || strings.HasPrefix(slashPath, match+"/") {
				return true
			}
		} else if strings.HasPrefix(ex, "**/") {
			match := strings.TrimPrefix(ex, "**/")
			matched, _ := filepath.Match(match, filepath.Base(slashPath))
			if matched {
				return true
			}
		} else if strings.HasSuffix(ex, "/") {
			match := strings.TrimSuffix(ex, "/")
			if strings.Contains(slashPath, "/"+match+"/") || strings.HasPrefix(slashPath, match+"/") || slashPath == match {
				return true
			}
		} else {
			matched, _ := filepath.Match(ex, filepath.Base(slashPath))
			if matched {
				return true
			}
		}
	}
	return false
}
"""
if "func isExcluded" not in text:
    text = text + is_ex_func

with open(path, "w") as f: f.write(text)
'
go build ./... && git add -A && git commit -m "update"
python3 -c '
path = "cmd/root.go"
with open(path, "r") as f: text = f.read()
if "\"archscan/internal/config\"" not in text:
    text = text.replace("\"archscan/internal/analyzer\"", "\"archscan/internal/analyzer\"\n\t\"archscan/internal/config\"")
if "cfg := config.Load(repoPath)" not in text:
    text = text.replace("result, err := analyzer.Analyze(repoPath, verbose)", "cfg := config.Load(repoPath)\n\tresult, err := analyzer.Analyze(repoPath, verbose, cfg)")
with open(path, "w") as f: f.write(text)
'
go build ./... && git add -A && git commit -m "update2"
