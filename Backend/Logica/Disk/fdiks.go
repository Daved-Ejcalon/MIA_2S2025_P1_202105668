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

// Fdisk crea particiones en un disco virtual existente (primarias, extendidas o lógicas)
func Fdisk(size int64, unit string, fit string, path string, ptype string, name string) error {
	unit = strings.ToUpper(unit)
	fit = strings.ToUpper(fit)
	ptype = strings.ToUpper(ptype)

	// Validaciones básicas de entrada
	if size <= 0 {
		return fmt.Errorf("tamaño inválido")
	}
	if name == "" {
		return errors.New("nombre requerido")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("nombre inválido")
	}

	// Conversión de unidades a bytes (K por defecto)
	if unit == "" {
		unit = "K"
	}
	switch unit {
	case "B":
	case "K":
		size *= 1024
	case "M":
		size *= 1024 * 1024
	default:
		return fmt.Errorf("unidad inválida")
	}

	if fit == "" {
		fit = "WF"
	}
	switch fit {
	case "BF", "FF", "WF":
	default:
		return fmt.Errorf("fit inválido")
	}

	if ptype == "" {
		ptype = "P"
	}
	switch ptype {
	case "P", "E", "L":
	default:
		return fmt.Errorf("tipo inválido")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe")
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Verificar nombres duplicados
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			return fmt.Errorf("nombre duplicado")
		}
	}

	// Contar particiones existentes y encontrar partición extendida
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

	// Validar límites de particiones según tipo (máximo 4 particiones en MBR)
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

	// Buscar slot disponible en la tabla de particiones
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

	// Manejar particiones lógicas usando EBR Manager
	if ptype == "L" {
		ebrMgr := Partition.NewEBRManager(path, extended)

		// Inicializar primer EBR si es necesario
		if err := ebrMgr.CreateFirstEBR(); err != nil {
			return fmt.Errorf("error inicializando EBR: %v", err)
		}

		if err := ebrMgr.AddLogicalPartition(name, size, fitToByte(fit)); err != nil {
			return fmt.Errorf("error creando partición lógica: %v", err)
		}

		return nil
	}

	// Calcular posición de inicio para particiones primarias/extendidas
	startPosition := int64(Models.GetMBRSize())
	
	// Encontrar la posición donde termina la última partición
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 {
			endPosition := partition.PartStart + partition.PartSize
			if endPosition > startPosition {
				startPosition = endPosition
			}
		}
	}

	// Verificar que hay espacio suficiente
	if startPosition+size > mbr.MbrSize {
		return fmt.Errorf("espacio insuficiente en el disco")
	}

	// Crear y escribir nueva partición en el MBR
	newPartition := Models.Partition{
		PartStatus:      1,
		PartType:        ptype[0],
		PartFit:         fitToByte(fit),
		PartStart:       startPosition,
		PartSize:        size,
		PartCorrelative: -1,
	}
	newPartition.SetPartitionName(name)

	mbr.Partitions[slotIndex] = newPartition

	file.Seek(0, 0)
	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		return fmt.Errorf("error actualizando MBR")
	}

	return nil
}

// fitToByte convierte el algoritmo de ajuste a byte para almacenamiento
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
