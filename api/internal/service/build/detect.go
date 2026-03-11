package build

import (
	"os"
	"path/filepath"
)

// DetectRuntime inspects the project directory to determine the application runtime.
// It checks for well-known files in priority order and returns a runtime name string.
func DetectRuntime(projectDir string) (string, error) {
	checks := []struct {
		file    string
		runtime string
	}{
		{"Dockerfile", "docker"},
		{"docker-compose.yml", "docker"},
		{"docker-compose.yaml", "docker"},
		{"package.json", "nodejs"},
		{"requirements.txt", "python"},
		{"Pipfile", "python"},
		{"go.mod", "go"},
		{"Gemfile", "ruby"},
		{"pom.xml", "java"},
		{"build.gradle", "java"},
		{"mix.exs", "elixir"},
		{"Cargo.toml", "rust"},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(projectDir, c.file)); err == nil {
			return c.runtime, nil
		}
	}

	// Check for .csproj files via glob
	matches, err := filepath.Glob(filepath.Join(projectDir, "*.csproj"))
	if err == nil && len(matches) > 0 {
		return "dotnet", nil
	}

	return "unknown", nil
}
