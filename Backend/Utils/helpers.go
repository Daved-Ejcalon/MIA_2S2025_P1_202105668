package Utils

import (
	"path/filepath"
)

// GetDirectory retorna el directorio de una ruta
func GetDirectory(path string) string {
	return filepath.Dir(path)
}

// GetFilename retorna solo el nombre del archivo
func GetFilename(path string) string {
	return filepath.Base(path)
}

// ConvertToBytes convierte tamaño y unidad a bytes
func ConvertToBytes(size int64, unit string) int64 {
	switch unit {
	case "K":
		return size * 1024
	case "M":
		return size * 1024 * 1024
	case "B":
		return size
	default:
		return size * 1024 * 1024 // Default a MB
	}
}
