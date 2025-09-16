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

// CatWithSession lee archivos usando un mountID específico de la sesión activa
func CatWithSession(fileArgs map[string]string, mountID string) error {
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

		err = readAndDisplayFileWithMountID(filePath, mountID)
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

// readAndDisplayFile lee contenido desde EXT2 usando la primera partición montada
func readAndDisplayFile(filePath string) error {
	// Obtener la primera partición montada disponible
	mountedPartitions := GetMountedPartitions()
	if len(mountedPartitions) == 0 {
		return fmt.Errorf("no hay particiones montadas")
	}

	// Usar la primera partición montada como predeterminada
	mountInfo := &mountedPartitions[0]

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
	err := ext2Manager.LoadPartitionInfo()
	if err != nil {
		return err
	}

	err = ext2Manager.LoadSuperBlock()
	if err != nil {
		return err
	}

	// Lee contenido del archivo usando el sistema de archivos EXT2
	fileManager := System.NewEXT2FileManager(ext2Manager)
	content, err := fileManager.ReadFileContent(filePath)
	if err != nil {
		return err
	}

	fmt.Print(content)
	return nil
}

// readAndDisplayFileWithMountID lee contenido desde EXT2 usando un mountID específico
func readAndDisplayFileWithMountID(filePath string, mountID string) error {
	mountInfo := findMountInfoByID(mountID)
	if mountInfo == nil {
		return fmt.Errorf("partición '%s' no está montada", mountID)
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
	err := ext2Manager.LoadPartitionInfo()
	if err != nil {
		return err
	}

	err = ext2Manager.LoadSuperBlock()
	if err != nil {
		return err
	}

	// Lee contenido del archivo usando el sistema de archivos EXT2
	fileManager := System.NewEXT2FileManager(ext2Manager)
	content, err := fileManager.ReadFileContent(filePath)
	if err != nil {
		return err
	}

	fmt.Print(content)
	return nil
}

// isSessionActive determina si el sistema permite operaciones de archivo
func isSessionActive() bool {
	// Verificar si hay particiones montadas como indicador de sesión activa
	return len(GetMountedPartitions()) > 0
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
