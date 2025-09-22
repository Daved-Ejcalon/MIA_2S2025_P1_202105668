package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
)

// MountInfo almacena información de particiones montadas
type MountInfo struct {
	DiskPath      string
	PartitionName string
	MountID       string
	DiskLetter    rune
	PartNumber    int
}

// Variables globales para el sistema de montaje
var (
	mountedPartitions   []MountInfo           // Lista de particiones montadas
	diskLetterMap       map[string]rune       // Mapeo de disco a letra asignada
	diskPartitionCount  map[string]int        // Contador de particiones por disco
	nextAvailableLetter rune            = 'A' // Siguiente letra disponible
)

// initMountSystem inicializa los mapas del sistema de montaje
func initMountSystem() {
	if diskLetterMap == nil {
		diskLetterMap = make(map[string]rune)
	}
	if diskPartitionCount == nil {
		diskPartitionCount = make(map[string]int)
	}
}

// Mount monta una partición y le asigna un ID único del formato {carnet}{num}{letra}
func Mount(path string, name string) error {
	initMountSystem()

	// Validaciones de entrada
	if path == "" {
		return fmt.Errorf("path requerido")
	}
	if name == "" {
		return fmt.Errorf("nombre requerido")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe")
	}

	// Verificar si la partición ya está montada
	if isAlreadyMounted(path, name) {
		return fmt.Errorf("partición ya montada")
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

	// Buscar la partición por nombre en el MBR (primarias y extendidas)
	var targetPartition *Models.Partition

	// Primero buscar en particiones primarias/extendidas del MBR
	for i, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			targetPartition = &mbr.Partitions[i]
			break
		}
	}

	// Si no se encontró, buscar en particiones lógicas
	if targetPartition == nil {
		targetPartition, _ = findLogicalPartition(file, &mbr, name)
	}

	if targetPartition == nil {
		return fmt.Errorf("partición no encontrada")
	}

	// Permitir montaje de particiones primarias, extendidas y lógicas
	// Las particiones extendidas se pueden montar para generar reportes EBR

	// Asignar letra de disco (reutilizar si ya existe, crear nueva si no)
	var diskLetter rune
	if letter, exists := diskLetterMap[path]; exists {
		diskLetter = letter
	} else {
		diskLetter = nextAvailableLetter
		diskLetterMap[path] = diskLetter
		diskPartitionCount[path] = 0
		nextAvailableLetter++
	}

	// Generar ID único y registrar el montaje
	diskPartitionCount[path]++
	partitionNumber := diskPartitionCount[path]

	mountID := fmt.Sprintf("68%d%c", partitionNumber, diskLetter)
	targetPartition.PartStatus = 1
	targetPartition.PartCorrelative = int64(partitionNumber)
	copy(targetPartition.PartID[:], mountID)

	// Escribir los cambios de vuelta al disco
	if targetPartition.PartType == 'L' {
		// Para particiones lógicas, actualizar el EBR correspondiente
		err = updateLogicalPartitionEBR(file, &mbr, name, mountID, partitionNumber)
		if err != nil {
			return fmt.Errorf("error actualizando EBR: %v", err)
		}
	} else {
		// Para particiones primarias y extendidas, actualizar el MBR
		for i := range mbr.Partitions {
			if mbr.Partitions[i].GetName() == name {
				mbr.Partitions[i] = *targetPartition
				break
			}
		}

		// Escribir MBR actualizado al disco
		file.Seek(0, 0)
		err = binary.Write(file, binary.LittleEndian, &mbr)
		if err != nil {
			return fmt.Errorf("error escribiendo MBR actualizado: %v", err)
		}
	}

	mountInfo := MountInfo{
		DiskPath:      path,
		PartitionName: name,
		MountID:       mountID,
		DiskLetter:    diskLetter,
		PartNumber:    partitionNumber,
	}
	mountedPartitions = append(mountedPartitions, mountInfo)

	return nil
}

// isAlreadyMounted verifica si una partición ya está montada
func isAlreadyMounted(path string, name string) bool {
	for _, mount := range mountedPartitions {
		if mount.DiskPath == path && mount.PartitionName == name {
			return true
		}
	}
	return false
}

// GetMountedPartitions retorna la lista de particiones montadas
func GetMountedPartitions() []MountInfo {
	return mountedPartitions
}

// UnmountPartition desmonta una partición por su ID de montaje
func UnmountPartition(mountID string) error {
	initMountSystem()

	// Buscar y remover la partición de la lista
	for i, mount := range mountedPartitions {
		if mount.MountID == mountID {
			mountedPartitions = append(mountedPartitions[:i], mountedPartitions[i+1:]...)
			diskPartitionCount[mount.DiskPath]--
			// Limpiar mapas si no quedan particiones del disco
			if diskPartitionCount[mount.DiskPath] == 0 {
				delete(diskLetterMap, mount.DiskPath)
				delete(diskPartitionCount, mount.DiskPath)
			}

			return nil
		}
	}

	return fmt.Errorf("ID no encontrado")
}

// ShowMountedPartitions muestra las particiones montadas (implementación pendiente)
func ShowMountedPartitions() {
	initMountSystem()

	if len(mountedPartitions) == 0 {
		return
	}

	for _, mount := range mountedPartitions {
		_ = mount
	}
}

// GetMountInfoByID obtiene información de montaje por ID
func GetMountInfoByID(mountID string) (*MountInfo, error) {
	initMountSystem()

	for _, mount := range mountedPartitions {
		if mount.MountID == mountID {
			return &mount, nil
		}
	}

	return nil, fmt.Errorf("partición con ID '%s' no está montada", mountID)
}

// GetMountedPartitionByID obtiene información de montaje por ID (para reportes)
func GetMountedPartitionByID(mountID string) *MountInfo {
	mountInfo, err := GetMountInfoByID(mountID)
	if err != nil {
		return nil
	}
	return mountInfo
}

// findLogicalPartition busca una partición lógica por nombre en las particiones extendidas
func findLogicalPartition(file *os.File, mbr *Models.MBR, name string) (*Models.Partition, bool) {
	// Buscar en cada partición extendida
	for _, partition := range mbr.Partitions {
		if partition.PartType == 'E' && partition.PartStatus != 0 {
			// Buscar en las particiones lógicas de esta extendida
			currentEBRPos := partition.PartStart

			for currentEBRPos != Models.EBR_END {
				// Leer EBR
				file.Seek(currentEBRPos, 0)
				var ebr Models.EBR
				err := binary.Read(file, binary.LittleEndian, &ebr)
				if err != nil {
					break
				}

				// Verificar si es la partición que buscamos
				if ebr.GetLogicalPartitionName() == name && ebr.PartS > 0 {
					// Crear una partición temporal con los datos del EBR
					logicalPartition := &Models.Partition{
						PartStatus: byte(ebr.PartMount),
						PartType:   'L', // Lógica
						PartFit:    ebr.PartFit,
						PartStart:  ebr.PartStart,
						PartSize:   ebr.PartS,
					}
					// Copiar el nombre
					copy(logicalPartition.PartName[:], ebr.PartName[:])
					return logicalPartition, true
				}

				// Siguiente EBR en la cadena
				if ebr.PartNext == Models.EBR_END {
					break
				}
				currentEBRPos = ebr.PartNext
			}
		}
	}

	return nil, false
}

// updateLogicalPartitionEBR actualiza el EBR de una partición lógica con información de montaje
func updateLogicalPartitionEBR(file *os.File, mbr *Models.MBR, name string, mountID string, partitionNumber int) error {
	// Buscar en cada partición extendida
	for _, partition := range mbr.Partitions {
		if partition.PartType == 'E' && partition.PartStatus != 0 {
			currentEBRPos := partition.PartStart

			for currentEBRPos != Models.EBR_END {
				// Leer EBR
				file.Seek(currentEBRPos, 0)
				var ebr Models.EBR
				err := binary.Read(file, binary.LittleEndian, &ebr)
				if err != nil {
					break
				}

				// Verificar si es la partición que buscamos
				if ebr.GetLogicalPartitionName() == name && ebr.PartS > 0 {
					// Actualizar el EBR con información de montaje
					ebr.PartMount = 1 // Marcar como montada
					// Escribir EBR actualizado
					file.Seek(currentEBRPos, 0)
					err = binary.Write(file, binary.LittleEndian, &ebr)
					if err != nil {
						return fmt.Errorf("error escribiendo EBR actualizado: %v", err)
					}
					return nil
				}

				// Siguiente EBR en la cadena
				if ebr.PartNext == Models.EBR_END {
					break
				}
				currentEBRPos = ebr.PartNext
			}
		}
	}

	return fmt.Errorf("no se pudo encontrar EBR para partición lógica %s", name)
}
