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

	// Buscar la partición por nombre en el MBR
	var targetPartition *Models.Partition

	for i, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			targetPartition = &mbr.Partitions[i]
			break
		}
	}

	if targetPartition == nil {
		return fmt.Errorf("partición no encontrada")
	}

	if targetPartition.PartType != 'P' {
		return fmt.Errorf("solo se pueden montar particiones primarias")
	}

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
