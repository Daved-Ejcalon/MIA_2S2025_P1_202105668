package Disk

import (
	"fmt"
)

func ShowDisk(diskArgs map[string]string) error {
	path, exists := diskArgs["path"]
	if !exists || path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	fmt.Printf("Analizando disco: %s\n", path)
	fmt.Println("[PENDIENTE] Lectura de MBR y tabla de particiones")

	return nil
}
