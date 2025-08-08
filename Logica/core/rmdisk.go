package core

import (
	"fmt"
	"os"
)

// RmDisk elimina un archivo .mia del disco
func RmDisk(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("El archivo %s no existe", path)
	}

	fmt.Printf("¿Está seguro que desea eliminar el disco en %s? (s/n): ", path)
	var input string
	fmt.Scanln(&input)

	if input != "s" && input != "S" {
		fmt.Println("Cancelado.")
		return nil
	}

	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("Error al eliminar el archivo: %v", err)
	}

	fmt.Println("Disco eliminado correctamente.")
	return nil
}
