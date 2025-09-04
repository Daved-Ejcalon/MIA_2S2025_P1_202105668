package Utils

import (
	"path/filepath"
)

func GetDirectory(path string) string {
	return filepath.Dir(path)
}

func GetFilename(path string) string {
	return filepath.Base(path)
}

func ConvertToBytes(size int64, unit string) int64 {
	switch unit {
	case "K":
		return size * 1024
	case "M":
		return size * 1024 * 1024
	case "B":
		return size
	default:
		return size * 1024 * 1024
	}
}
