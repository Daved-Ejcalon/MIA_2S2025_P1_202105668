package utils

import "path/filepath"

func GetDir(path string) string {
	return filepath.Dir(path)
}
