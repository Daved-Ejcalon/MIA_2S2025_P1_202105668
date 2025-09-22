package Reportes

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes/Graphviz"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateEBRReport genera un reporte del EBR en formato JPG usando Graphviz
func GenerateEBRReport(partitionID string, ebrName string, outputPath string) error {
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

	// Validar que se proporcionó el nombre del EBR
	if ebrName == "" {
		return fmt.Errorf("debe especificar el nombre del EBR a reportar")
	}

	// Crear directorio de salida si no existe
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %v", err)
	}

	// Generar el reporte usando Graphviz
	err := Graphviz.GenerateEBRGraph(diskPath, ebrName, outputPath)
	if err != nil {
		return fmt.Errorf("error generando reporte EBR: %v", err)
	}

	fmt.Println("Reporte generado exitosamente")
	return nil
}

// GenerateEBRCompleteReport genera un reporte completo de todos los EBRs en formato JPG
func GenerateEBRCompleteReport(partitionID string, outputPath string) error {
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

	// Generar el reporte completo usando Graphviz
	err := Graphviz.GenerateEBRCompleteGraph(diskPath, outputPath)
	if err != nil {
		return fmt.Errorf("error generando reporte EBR completo: %v", err)
	}

	fmt.Println("Reporte generado exitosamente")
	return nil
}