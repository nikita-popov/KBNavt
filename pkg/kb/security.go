package kb

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SecurityManager handles path validation and sandboxing
type SecurityManager struct {
	AllowedRoots []string
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(roots ...string) *SecurityManager {
	return &SecurityManager{
		AllowedRoots: roots,
	}
}

// ValidatePath ensures a path is within allowed roots
func (sm *SecurityManager) ValidatePath(requestPath string) (string, error) {
	// Clean the path to prevent traversal
	cleanPath := filepath.Clean(requestPath)

	// Remove leading slashes for relative path comparison
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// Check if path attempts directory traversal
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("path traversal detected: %s", requestPath)
	}

	// Try to resolve within each allowed root
	for _, root := range sm.AllowedRoots {
		fullPath := filepath.Join(root, cleanPath)
		fullPath = filepath.Clean(fullPath)
		root = filepath.Clean(root)

		// Verify the resolved path is still within root
		rel, err := filepath.Rel(root, fullPath)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("path not in allowed roots: %s", requestPath)
}

// IsAllowedFile checks if file extension is allowed
func (sm *SecurityManager) IsAllowedFile(filename string) bool {
	allowed := map[string]bool{
		".org":  true,
		".md":   true,
		".markdown": true,
		".txt":  true,
	}

	ext := strings.ToLower(filepath.Ext(filename))
	return allowed[ext]
}
