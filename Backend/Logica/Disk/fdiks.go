package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Fdisk crea una partición en el disco especificado
func Fdisk(size int64, unit string, fit string, path string, ptype string, name string) error {
	// Normalizar parámetros
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)
	ptype = strings.ToUpper(ptype)

	// Validaciones básicas con mensajes específicos
	if size <= 0 {
		return fmt.Errorf("error: el tamaño de la partición debe ser mayor a 0, se proporcionó: %d", size)
	}
	if name == "" {
		return errors.New("error: el nombre de la partición es obligatorio y no puede estar vacío")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("error: el nombre de la partición no puede contener solo espacios en blanco")
	}

	// Convertir unidades (default K) con mensajes específicos
	if unit == "" {
		unit = "K"
	}
	originalSize := size
	switch unit {
	case "B":
		// No cambiar
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return fmt.Errorf("error: unidad '%s' no válida. Las unidades permitidas son: B (bytes), K (kilobytes), M (megabytes)", unit)
	}

	// Validar fit (default WF) con mensaje específico
	if fit == "" {
		fit = "WF"
	}
	switch fit {
	case "BF", "FF", "WF":
		// Válido
	default:
		return fmt.Errorf("error: tipo de ajuste '%s' no válido. Los ajustes permitidos son: BF (Best Fit), FF (First Fit), WF (Worst Fit)", fit)
	}

	// Validar tipo (default P) con mensaje específico
	if ptype == "" {
		ptype = "P"
	}
	switch ptype {
	case "P", "E", "L":
		// Válido
	default:
		return fmt.Errorf("error: tipo de partición '%s' no válido. Los tipos permitidos son: P (primaria), E (extendida), L (lógica)", ptype)
	}

	// Verificar que el archivo existe antes de intentar abrirlo
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("error: el archivo de disco '%s' no existe. Verifique la ruta o cree el disco primero con mkdisk", path)
	}

	// Abrir archivo del disco
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error: no se pudo abrir el archivo de disco '%s': %v", path, err)
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error: no se pudo leer el MBR del disco '%s'. El archivo puede estar corrupto: %v", path, err)
	}

	// Verificar nombre único con mensaje específico
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			return fmt.Errorf("error: ya existe una partición con el nombre '%s' en el disco '%s'. Los nombres de partición deben ser únicos", name, path)
		}
	}

	// Contar particiones existentes
	var primaryCount, extendedCount int
	var extended *Models.Partition
	for i := range mbr.Partitions {
		p := &mbr.Partitions[i]
		if p.PartStatus != 0 {
			switch p.PartType {
			case 'P':
				primaryCount++
			case 'E':
				extendedCount++
				extended = p
			}
		}
	}

	// Validaciones específicas según tipo de partición
	if ptype == "P" && primaryCount >= 4 {
		return fmt.Errorf("error: no se puede crear la partición primaria '%s'. Ya existen 4 particiones primarias (máximo permitido)", name)
	}

	if ptype == "P" && primaryCount >= 3 && extendedCount == 1 {
		return fmt.Errorf("error: no se puede crear la partición primaria '%s'. Ya existen 3 particiones primarias y 1 extendida (máximo: 3P + 1E)", name)
	}

	if ptype == "E" && extendedCount >= 1 {
		return fmt.Errorf("error: no se puede crear la partición extendida '%s'. Ya existe una partición extendida en el disco (solo se permite una por disco)", name)
	}

	if ptype == "L" && extendedCount == 0 {
		return fmt.Errorf("error: no se puede crear la partición lógica '%s'. No existe una partición extendida en el disco (las particiones lógicas requieren una partición extendida)", name)
	}

	// Buscar slot libre en el arreglo del MBR
	slotIndex := -1
	for i, partition := range mbr.Partitions {
		if partition.PartStatus == 0 {
			slotIndex = i
			break
		}
	}
	if slotIndex == -1 {
		return fmt.Errorf("error: no se puede crear la partición '%s'. No hay espacios disponibles en la tabla de particiones del disco", name)
	}

	// Calcular posición de inicio
	startPosition := int64(binary.Size(mbr))

	// Lógica específica para particiones lógicas
	if ptype == "L" {
		startPosition = extended.PartStart

		// Encontrar espacio después de todas las lógicas existentes
		for _, p := range mbr.Partitions {
			if p.PartStatus != 0 && p.PartType == 'L' {
				end := p.PartStart + p.PartSize
				if end > startPosition {
					startPosition = end
				}
			}
		}

		// Verificar que cabe dentro de la extendida
		extEnd := extended.PartStart + extended.PartSize
		if startPosition+size > extEnd {
			availableSpace := extEnd - startPosition
			return fmt.Errorf("error: no se puede crear la partición lógica '%s' de %d %s (%d bytes). Espacio disponible en la partición extendida: %d bytes",
				name, originalSize, unit, size, availableSpace)
		}
	} else {
		// Para particiones primarias y extendidas
		for _, partition := range mbr.Partitions {
			if partition.PartStatus != 0 {
				endPosition := partition.PartStart + partition.PartSize
				if endPosition > startPosition {
					startPosition = endPosition
				}
			}
		}

		// Verificar espacio en el disco
		if startPosition+size > mbr.MbrSize {
			availableSpace := mbr.MbrSize - startPosition
			return fmt.Errorf("error: no se puede crear la partición '%s' de %d %s (%d bytes). Espacio disponible en el disco: %d bytes",
				name, originalSize, unit, size, availableSpace)
		}
	}

	// Crear nueva partición
	newPartition := Models.Partition{
		PartStatus:      1,
		PartType:        ptype[0],
		PartFit:         fit[0],
		PartStart:       startPosition,
		PartSize:        size,
		PartCorrelative: -1,
	}
	newPartition.SetPartitionName(name)

	// Guardar en el slot disponible
	mbr.Partitions[slotIndex] = newPartition

	// Escribir MBR actualizado
	file.Seek(0, 0)
	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		return fmt.Errorf("error: no se pudo actualizar el MBR del disco '%s': %v", path, err)
	}

	// Mensaje de éxito detallado
	fmt.Printf("✓ Partición '%s' creada exitosamente\n", name)
	fmt.Printf("  Tipo: %s", ptype)
	if ptype == "P" {
		fmt.Printf(" (Primaria)")
	} else if ptype == "E" {
		fmt.Printf(" (Extendida)")
	} else if ptype == "L" {
		fmt.Printf(" (Lógica)")
	}
	fmt.Printf("\n  Tamaño: %d %s (%d bytes)\n", originalSize, unit, size)
	fmt.Printf("  Ajuste: %s", fit)
	if fit == "BF" {
		fmt.Printf(" (Best Fit)")
	} else if fit == "FF" {
		fmt.Printf(" (First Fit)")
	} else if fit == "WF" {
		fmt.Printf(" (Worst Fit)")
	}
	fmt.Printf("\n  Posición: %d\n", startPosition)
	fmt.Printf("  Disco: %s\n", path)

	return nil
}
