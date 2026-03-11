package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var DefaultExcludes = []string{
	".git",
	"node_modules",
	"__pycache__",
	".venv",
	"vendor",
	".deployer.json",
	".DS_Store",
	"*.log",
}

// ZipDirectory creates a temporary zip file from sourceDir, excluding files
// matching the given patterns. Returns the path to the created zip file.
func ZipDirectory(sourceDir string, excludePatterns []string) (string, error) {
	if len(excludePatterns) == 0 {
		excludePatterns = DefaultExcludes
	}

	tmpFile, err := os.CreateTemp("", "deployer-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	writer := zip.NewWriter(tmpFile)
	defer writer.Close()

	sourceDir, err = filepath.Abs(sourceDir)
	if err != nil {
		return "", err
	}

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if shouldExclude(relPath, info.IsDir(), excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	})

	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func shouldExclude(relPath string, isDir bool, patterns []string) bool {
	base := filepath.Base(relPath)
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	for _, pattern := range patterns {
		// Check if any path component matches the pattern
		for _, part := range parts {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
		// Also check the base name
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}
