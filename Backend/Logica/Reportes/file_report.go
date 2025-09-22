package Reportes

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes/Graphviz"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateFileReport genera un reporte del contenido de un archivo específico
func GenerateFileReport(partitionID string, outputPath string, pathFileLS string) error {
	// Validar que se proporcione la ruta del archivo
	if pathFileLS == "" {
		return fmt.Errorf("debe especificar la ruta del archivo con -path_file_ls")
	}

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
	_, _, err := Users.GetPartitionAndSuperBlock(mountedPartition)
	if err != nil {
		return fmt.Errorf("error obteniendo superbloque: %v", err)
	}

	// Convertir MountInfo de Disk a System
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountedPartition.DiskPath,
		PartitionName: mountedPartition.PartitionName,
		MountID:       mountedPartition.MountID,
		DiskLetter:    mountedPartition.DiskLetter,
		PartNumber:    mountedPartition.PartNumber,
	}

	// Crear EXT2Manager
	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("error creando EXT2Manager")
	}

	// Crear FileManager
	fileManager := System.NewEXT2FileManager(ext2Manager)
	if fileManager == nil {
		return fmt.Errorf("error creando FileManager")
	}

	// Leer el contenido del archivo
	content, err := fileManager.ReadFileContent(pathFileLS)
	if err != nil {
		return fmt.Errorf("error leyendo archivo '%s': %v", pathFileLS, err)
	}

	// Crear directorio de salida si no existe
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %v", err)
	}

	// Obtener nombre del disco y archivo
	diskName := filepath.Base(diskPath)
	fileName := filepath.Base(pathFileLS)

	// Generar el reporte usando Graphviz
	err = Graphviz.GenerateFileGraph(fileName, pathFileLS, content, diskName, outputPath)
	if err != nil {
		return fmt.Errorf("error generando reporte de archivo: %v", err)
	}

	fmt.Println("Reporte de archivo generado exitosamente")
	return nil
}
