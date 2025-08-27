package Disk

import (
	"fmt"
	"os"
)

// RmDisk elimina un archivo de disco del sistema
func RmDisk(path string) error {
	// Eliminar archivo
	os.Remove(path)

	fmt.Printf("Disco eliminado\n")
	return nil
}
