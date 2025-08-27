package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/Partition"
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Fdisk, para crear una partición en el disco especificado
func Fdisk(size int64, unit string, fit string, path string, ptype string, name string) error {
	// Normalizar parámetros
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)
	ptype = strings.ToUpper(ptype)

	// Validaciones básicas con mensajes específicos
	if size <= 0 {
		return fmt.Errorf("tamaño inválido")
	}
	if name == "" {
		return errors.New("nombre requerido")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("nombre inválido")
	}

	// Convertir unidades (Por default K)
	if unit == "" {
		unit = "K"
	}
	switch unit {
	case "B":
		// No cambiar
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return fmt.Errorf("unidad inválida")
	}

	// Validar fit (Por default WF)
	if fit == "" {
		fit = "WF"
	}
	switch fit {
	case "BF", "FF", "WF":
		// Válido
	default:
		return fmt.Errorf("fit inválido")
	}

	// Validar tipo (default P)
	if ptype == "" {
		ptype = "P"
	}
	switch ptype {
	case "P", "E", "L":
	// Válido
	default:
		return fmt.Errorf("tipo inválido")
	}

	// Verificar que el archivo existe antes de intentar abrirlo
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe")
	}

	// Abrir archivo del disco
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Verificar nombre único con mensaje específico
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			return fmt.Errorf("nombre duplicado")
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
		return fmt.Errorf("demasiadas particiones primarias")
	}

	if ptype == "P" && primaryCount >= 3 && extendedCount == 1 {
		return fmt.Errorf("límite de particiones alcanzado")
	}

	if ptype == "E" && extendedCount >= 1 {
		return fmt.Errorf("partición extendida existente")
	}

	if ptype == "L" && extendedCount == 0 {
		return fmt.Errorf("partición extendida requerida")
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
		return fmt.Errorf("sin espacios disponibles")
	}

	// Calcular posición de inicio
	startPosition := int64(binary.Size(mbr))

	// Lógica específica para particiones lógicas
	if ptype == "L" {
		// Para particiones lógicas, usar EBR_Manager
		ebrMgr := Partition.NewEBRManager(path, extended)

		// Convertir tamaño a bytes
		sizeInBytes := size
		if unit == "K" {
			sizeInBytes *= 1024
		} else if unit == "M" {
			sizeInBytes *= 1024 * 1024
		}

		// Crear partición lógica usando EBR_Manager
		if err := ebrMgr.AddLogicalPartition(name, sizeInBytes, fitToByte(fit)); err != nil {
			return fmt.Errorf("error creando partición lógica")
		}

		fmt.Printf("Partición lógica '%s' creada\n", name)
		return nil
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
			return fmt.Errorf("espacio insuficiente")
		}
	}

	// Crear nueva partición
	newPartition := Models.Partition{
		PartStatus:      1,
		PartType:        ptype[0],
		PartFit:         fitToByte(fit),
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
		return fmt.Errorf("error actualizando MBR")
	}

	fmt.Printf("Partición '%s' creada\n", name)

	return nil
}

// fitToByte convierte string de fit a byte
func fitToByte(fit string) byte {
	s := strings.ToUpper(fit)
	if s == "BF" || s == "B" {
		return Models.FIT_BEST
	}
	if s == "FF" || s == "F" {
		return Models.FIT_FIRST
	}
	return Models.FIT_WORST
}
