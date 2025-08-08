package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"

	"Logica/models"
)

func Fdisk(size int64, unit string, fit string, path string, ptype string, name string) error {
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)
	ptype = strings.ToUpper(ptype)

	if size <= 0 {
		return errors.New("El tamaño debe ser mayor a 0")
	}

	if unit == "" {
		unit = "K"
	}

	switch unit {
	case "B":
		// no cambiar
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return errors.New("Unidad no válida (B, K, M)")
	}

	if fit == "" {
		fit = "WF"
	}

	// Validación de tipo de ajuste (fit)
	switch fit {
	case "BF", "FF", "WF":
		// válido
	default:
		return errors.New("Tipo de ajuste inválido. Use BF (Best Fit), FF (First Fit) o WF (Worst Fit)")
	}

	if ptype == "" {
		ptype = "P"
	}

	// Validación de tipo de partición
	switch ptype {
	case "P", "E", "L":
		// válido
	default:
		return errors.New("Tipo de partición no válido. Use P (Primaria), E (Extendida) o L (Lógica)")
	}

	// Leer MBR
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("No se pudo abrir el archivo: %v", err)
	}
	defer file.Close()

	var mbr models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("Error al leer el MBR: %v", err)
	}

	// Validar nombre único y contar particiones existentes
	var primaryCount, extendedCount int
	for _, p := range mbr.Partitions {
		pname := strings.TrimSpace(string(p.PartName[:]))
		if pname == name {
			return errors.New("Ya existe una partición con ese nombre")
		}
		if p.PartStatus != 0 {
			switch p.PartType {
			case 'P':
				primaryCount++
			case 'E':
				extendedCount++
			}
		}
	}

	if ptype == "P" && primaryCount >= 3 && extendedCount == 1 {
		return errors.New("Ya hay 3 primarias y 1 extendida, no se puede crear otra")
	}
	if ptype == "E" && extendedCount >= 1 {
		return errors.New("Solo se permite una partición extendida por disco")
	}

	// Buscar slot disponible
	var index = -1
	for i, p := range mbr.Partitions {
		if p.PartStatus == 0 {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New("No hay espacio en el MBR para más particiones")
	}

	// Calcular inicio de la nueva partición
	var start int64 = int64(binary.Size(mbr))
	for _, p := range mbr.Partitions {
		if p.PartStatus != 0 && p.PartStart+int64(p.PartSize) > start {
			start = p.PartStart + int64(p.PartSize)
		}
	}

	// Crear la partición
	var part models.Partition
	part.PartStatus = 1
	part.PartType = ptype[0]
	part.PartFit = fit[0]
	part.PartStart = start
	part.PartSize = size
	part.PartCorrelative = -1

	copy(part.PartName[:], name)

	mbr.Partitions[index] = part

	// Reescribir MBR
	file.Seek(0, 0)
	err = binary.Write(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("Error al escribir el nuevo MBR: %v", err)
	}

	fmt.Println("Partición creada correctamente:", name)
	return nil
}
