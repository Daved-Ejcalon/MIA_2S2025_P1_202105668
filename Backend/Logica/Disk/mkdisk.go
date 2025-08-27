package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MkDisk crea un disco virtual con el tamaño especificado
func MkDisk(size int64, unit string, fit string, path string) error {
	// Normalizar parámetros
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)

	// Validaciones
	if size <= 0 {
		return errors.New("tamaño inválido")
	}


	// Validar unidad
	if unit == "" {
		unit = "M"
	}
	switch unit {
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	case "B":

	default:
		return errors.New("unidad inválida")
	}

	// Validar fit
	if fit == "" {
		fit = "FF"
	}
	switch fit {
	case "BF", "FF", "WF":
	// Válido
	default:
		return errors.New("fit inválido")
	}

	// Crear directorios si no existen
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creando directorios")
	}

	// Crear archivo
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creando archivo")
	}
	defer file.Close()

	// Llenar archivo con ceros usando buffer
	buffer := make([]byte, 1024)
	remaining := size

	for remaining > 0 {
		writeSize := int64(1024)
		if remaining < 1024 {
			writeSize = remaining
		}

		_, err := file.Write(buffer[:writeSize])
		if err != nil {
			return fmt.Errorf("error escribiendo archivo")
		}

		remaining -= writeSize
	}

	// Crear y escribir MBR
	mbr := Models.MBR{
		MbrSize:         size,
		MbrCreationDate: time.Now().Unix(),
		MbrSignature:    rand.Int63(),
		DiskFit:         fit[0],
	}

	// Inicializar particiones vacías
	for i := range mbr.Partitions {
		mbr.Partitions[i].PartCorrelative = -1
	}

	// Escribir MBR al inicio del archivo
	file.Seek(0, 0)
	err = binary.Write(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	fmt.Printf("Disco creado: %s\n", path)

	return nil
}
