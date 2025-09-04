package Disk

import (
	"errors"
	"os"
)

// RmDisk elimina un disco virtual del sistema de archivos
func RmDisk(path string) error {
	// Verificar que el archivo existe antes de intentar eliminarlo
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("el archivo no existe")
	}

	// Eliminar el archivo del disco
	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}
