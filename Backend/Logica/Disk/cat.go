package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/System"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Cat lee y muestra el contenido de archivos desde particiones EXT2 montadas
func Cat(fileArgs map[string]string) error {
	if !isSessionActive() {
		return fmt.Errorf("sesión requerida")
	}

	fileList, err := extractFileParameters(fileArgs)
	if err != nil {
		return err
	}

	if len(fileList) == 0 {
		return fmt.Errorf("archivos requeridos")
	}

	for _, filePath := range fileList {
		err := validateFileAccess(filePath)
		if err != nil {
			return err
		}

		err = readAndDisplayFile(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractFileParameters procesa parametros file1,2,3, etc
func extractFileParameters(args map[string]string) ([]string, error) {
	fileMap := make(map[int]string)

	for key, value := range args {
		if strings.HasPrefix(key, "file") {
			numStr := strings.TrimPrefix(key, "file")
			if numStr == "" {
				return nil, fmt.Errorf("parámetro inválido")
			}

			num, err := strconv.Atoi(numStr)
			if err != nil {
				return nil, fmt.Errorf("parámetro inválido")
			}

			if value == "" {
				return nil, fmt.Errorf("valor requerido")
			}

			fileMap[num] = value
		}
	}

	fileList := make([]string, 0, len(fileMap))

	// Ordena indices para procesar archivos secuencialmente
	indices := make([]int, 0, len(fileMap))
	for index := range fileMap {
		indices = append(indices, index)
	}
	sort.Ints(indices)

	for _, index := range indices {
		fileList = append(fileList, fileMap[index])
	}

	return fileList, nil
}

// validateFileAccess verificacion del path
func validateFileAccess(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("path inválida")
	}

	return nil
}

// readAndDisplayFile lee contenido desde EXT2 usando formato mountID:/path
func readAndDisplayFile(filePath string) error {
	mountID, actualPath, err := parseFilePath(filePath)
	if err != nil {
		return err
	}

	mountInfo := findMountInfoByID(mountID)
	if mountInfo == nil {
		return fmt.Errorf("partición no montada")
	}

	// Convierte estructura MountInfo entre paquetes para compatibilidad
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}
	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	// Inicializa sistema EXT2 cargando metadatos de partición
	err = ext2Manager.LoadPartitionInfo()
	if err != nil {
		return err
	}

	err = ext2Manager.LoadSuperBlock()
	if err != nil {
		return err
	}

	// Lee contenido del archivo usando el sistema de archivos EXT2
	fileManager := System.NewEXT2FileManager(ext2Manager)
	content, err := fileManager.ReadFileContent(actualPath)
	if err != nil {
		return err
	}

	fmt.Print(content)
	return nil
}

// isSessionActive determina si el sistema permite operaciones de archivo
func isSessionActive() bool {
	return true
}

// parseFilePath separa mountID:/path en componentes individuales
func parseFilePath(filePath string) (string, string, error) {
	parts := strings.Split(filePath, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("formato inválido, use mountID:/path")
	}
	return parts[0], parts[1], nil
}

// findMountInfoByID busca información de montaje por ID de partición
func findMountInfoByID(mountID string) *MountInfo {
	for _, mount := range GetMountedPartitions() {
		if mount.MountID == mountID {
			return &mount
		}
	}
	return nil
}
