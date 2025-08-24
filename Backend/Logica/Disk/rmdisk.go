package Disk

import (
	"fmt"
	"os"
)

// RmDisk elimina un archivo de disco del sistema
func RmDisk(path string) error {
	// Verificar que el archivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("el archivo %s no existe", path)
	}

	// Confirmar eliminación
	fmt.Printf("¿Está seguro que desea eliminar el disco '%s'? (s/n): ", path)
	var response string
	fmt.Scanln(&response)

	if response != "s" && response != "S" {
		fmt.Println("Operación cancelada")
		return nil
	}

	// Eliminar archivo
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("error eliminando archivo: %w", err)
	}

	fmt.Printf("Disco '%s' eliminado exitosamente\n", path)
	return nil
}
