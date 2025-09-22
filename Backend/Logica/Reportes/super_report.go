package Reportes

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes/Graphviz"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateSuperBlockReport genera un reporte del superbloque en formato JPG usando Graphviz
func GenerateSuperBlockReport(partitionID string, outputPath string) error {
	// Buscar la partición montada por ID
	mountedPartition := Disk.GetMountedPartitionByID(partitionID)
	if mountedPartition == nil {
		return fmt.Errorf("particion con ID '%s' no encontrada o no montada", partitionID)
	}

	// Verificar que el disco existe
	diskPath := mountedPartition.DiskPath
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Errorf("el disco '%s' no existe", diskPath)
	}

	// Obtener información de la partición y superbloque
	_, superblock, err := Users.GetPartitionAndSuperBlock(mountedPartition)
	if err != nil {
		return fmt.Errorf("error obteniendo superbloque: %v", err)
	}

	// Crear directorio de salida si no existe
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %v", err)
	}

	// Obtener nombre del disco para mostrar en el reporte
	diskName := filepath.Base(diskPath)

	// Generar el reporte usando Graphviz
	err = Graphviz.GenerateSuperBlockGraph(superblock, diskName, outputPath)
	if err != nil {
		return fmt.Errorf("error generando reporte de superbloque: %v", err)
	}

	fmt.Println("Reporte generado exitosamente")
	return nil
}