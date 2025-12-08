package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func validatePath(baseDir, userPath string) (string, error) {
    absBase, _ := filepath.Abs(baseDir)
    absReq, _ := filepath.Abs(filepath.Join(baseDir, userPath))

    if !strings.HasPrefix(absReq, absBase) {
        return "", fmt.Errorf("path escape")
    }

    return absReq, nil
}

func validateQuery(q string) error {
    if len(strings.TrimSpace(q)) == 0 {
        return fmt.Errorf("empty query")
    }
    if len(q) > 1000 {
        return fmt.Errorf("query too long")
    }
    if !utf8.ValidString(q) {
        return fmt.Errorf("invalid UTF-8")
    }
    return nil
}
