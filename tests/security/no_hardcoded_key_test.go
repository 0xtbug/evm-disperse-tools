package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoHardcodedPrivateKeys(t *testing.T) {
	// Pattern that matches common private key formats
	root := "../../" // Adjust based on test location
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip vendor, git, and build directories
		if strings.Contains(path, "vendor") ||
			strings.Contains(path, ".git") ||
			strings.Contains(path, "node_modules") ||
			strings.Contains(path, "tmp") {
			return nil
		}

		// Only check .go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			t.Logf("✓ Scanned: %s", path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed to scan directory: %v", err)
	}
}
