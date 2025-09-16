package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/System"
	"fmt"
)

// Mkfs formatea una partición montada con sistema de archivos EXT2
func Mkfs(mountID string) error {
	// Validar que el ID de montaje esté presente
	if mountID == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	// Buscar la partición montada por ID
	mountInfo, err := findMountedPartitionByID(mountID)
	if err != nil {
		return fmt.Errorf("particion no montada")
	}

	// Crear estructura de montaje para el sistema de archivos
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	// Inicializar EXT2 manager y formatear la partición
	ext2Manager := System.NewEXT2Manager(systemMountInfo)

	err = ext2Manager.FormatPartition()
	if err != nil {
		return fmt.Errorf("fallo el formateo: %v", err)
	}

	return nil
}

// findMountedPartitionByID busca una partición montada por su ID único
func findMountedPartitionByID(mountID string) (*MountInfo, error) {
	mountedPartitions := GetMountedPartitions()

	// Buscar en lista de particiones montadas
	for _, mount := range mountedPartitions {
		if mount.MountID == mountID {
			return &mount, nil
		}
	}

	return nil, fmt.Errorf("partición no encontrada")
}
