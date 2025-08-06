package core

import (
	"Logica/models"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func MkDisk(size int64, unit string, fit string, path string) error {
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)

	if size <= 0 {
		return errors.New("El parámetro size debe ser un número entero positivo")
	}
	if fit != "BF" && fit != "FF" && fit != "WF" {
		return fmt.Errorf("El parámetro fit debe ser uno de: BF, FF, WF")
	}
	if unit != "K" && unit != "M" && unit != "" {
		return fmt.Errorf("El parámetro unit debe ser K o M")
	}
	if unit == "" {
		unit = "M"
	}

	var finalSize int64
	switch unit {
	case "K":
		finalSize = size * 1024
	case "M":
		finalSize = size * 1024 * 1024
	}

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear carpetas: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error al crear archivo: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for i := int64(0); i < finalSize/1024; i++ {
		file.Write(buffer)
	}

	mbr := models.MBR{
		MbrSize:         finalSize,
		MbrCreationDate: time.Now(),
		MbrSignature:    rand.Int63(),
		DiskFit:         fit[0],
	}

	file.Seek(0, 0)
	err = binary.Write(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("Error al escribir MBR: %v", err)
	}

	fmt.Println("Disco creado correctamente en:", path)
	return nil
}
