package Disk

import "fmt"

// Mounted muestra en consola todas las particiones actualmente montadas
func Mounted() {
	// Verificar si hay particiones montadas
	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas")
		return
	}

	// Listar todas las particiones montadas con formato ID | Nombre -> Ruta
	for _, mount := range mountedPartitions {
		fmt.Printf("ID: %s | %s -> %s\n",
			mount.MountID,
			mount.PartitionName,
			mount.DiskPath)
	}
}
