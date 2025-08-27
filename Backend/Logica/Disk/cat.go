package Disk

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Se muestra el contenido de archivos del sistema EXT2
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
		fmt.Printf("=== %s ===\n", filePath)

		err := validateFileAccess(filePath)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		err = readAndDisplayFile(filePath)
		if err != nil {
			fmt.Printf("❌ Error leyendo archivo: %v\n", err)
			continue
		}
	}

	return nil
}

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
				return nil, fmt.Errorf("valor de archivo requerido")
			}

			fileMap[num] = value
		}
	}

	fileList := make([]string, 0, len(fileMap))

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

func validateFileAccess(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("ruta inválida")
	}

	return nil // TODO: Implementar validación EXT2
}

func readAndDisplayFile(filePath string) error {
	fmt.Printf("[Funcionalidad EXT2 en desarrollo]\n")

	return nil // TODO: Implementar lectura EXT2
}

func isSessionActive() bool {
	return false // TODO: Implementar sistema de sesiones
}
