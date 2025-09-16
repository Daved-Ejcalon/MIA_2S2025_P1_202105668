package Reportes

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes/Graphviz"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateDiskReport genera un reporte del disco en formato JPG usando Graphviz
func GenerateDiskReport(partitionID string, outputPath string) error {
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

	// Crear directorio de salida si no existe
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %v", err)
	}

	// Generar el reporte usando Graphviz
	err := Graphviz.GenerateDiskGraph(diskPath, outputPath)
	if err != nil {
		return fmt.Errorf("error generando reporte de disco: %v", err)
	}

	fmt.Printf("Reporte de disco generado exitosamente: %s\n", outputPath)
	return nil
}