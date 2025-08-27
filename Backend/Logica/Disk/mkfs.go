package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/System"
	"fmt"
	"strings"
)

// Mkfs, para formatear una partición montada con EXT2
func Mkfs(mountID string, fsType string, formatType string) error {
	if mountID == "" {
		return fmt.Errorf("parámetro -id requerido")
	}

	mountInfo, err := findMountedPartitionByID(mountID)
	if err != nil {
		return fmt.Errorf("partición no montada")
	}

	if formatType == "" {
		formatType = "Full"
	}
	formatType = strings.ToUpper(formatType)
	if formatType != "FULL" {
		return fmt.Errorf("tipo de formateo inválido")
	}

	if fsType == "" {
		fsType = "2fs"
	}
	if fsType != "2fs" && fsType != "3fs" {
		return fmt.Errorf("sistema de archivos no soportado")
	}

	fmt.Printf("Formateando partición %s como EXT2\n", mountID)

	// Convertir a System.MountInfo
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	// Inicializar EXT2Manager
	ext2Manager := System.NewEXT2Manager(systemMountInfo)

	// Ejecutar formateo
	err = ext2Manager.FormatPartition()
	if err != nil {
		return fmt.Errorf("falló el formateo")
	}

	fmt.Printf("Partición %s formateada como EXT2\n", mountID)

	return nil
}

// findMountedPartitionByID busca una partición montada por su ID
func findMountedPartitionByID(mountID string) (*MountInfo, error) {
	mountedPartitions := GetMountedPartitions()

	for _, mount := range mountedPartitions {
		if mount.MountID == mountID {
			return &mount, nil
		}
	}

	return nil, fmt.Errorf("partición no encontrada")
}
