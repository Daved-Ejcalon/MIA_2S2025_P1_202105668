package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
)

// MountInfo representa información de una partición montada
type MountInfo struct {
	DiskPath      string // Ruta del disco
	PartitionName string // Nombre de la partición
	MountID       string // ID asignado (ej: "681A")
	DiskLetter    rune   // Letra asignada al disco ('A', 'B', 'C'...)
	PartNumber    int    // Número de partición montada en ese disco
}

// Variables globales para manejo de montajes en RAM
var (
	mountedPartitions   []MountInfo           // Tabla de montajes activos
	diskLetterMap       map[string]rune       // Mapa: path -> letra asignada
	diskPartitionCount  map[string]int        // Mapa: path -> contador particiones
	nextAvailableLetter rune            = 'A' // Próxima letra disponible
)

// Inicializar variables globales si no existen
func initMountSystem() {
	if diskLetterMap == nil {
		diskLetterMap = make(map[string]rune)
	}
	if diskPartitionCount == nil {
		diskPartitionCount = make(map[string]int)
	}
}

// Mount monta una partición en el sistema
func Mount(path string, name string) error {
	initMountSystem()

	// Validaciones básicas
	if path == "" {
		return fmt.Errorf("error: la ruta del disco es obligatoria")
	}
	if name == "" {
		return fmt.Errorf("error: el nombre de la partición es obligatorio")
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("error: el archivo de disco '%s' no existe", path)
	}

	// Verificar si ya está montada
	if isAlreadyMounted(path, name) {
		return fmt.Errorf("error: la partición '%s' del disco '%s' ya está montada", name, path)
	}

	// Abrir archivo del disco
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error: no se pudo abrir el disco '%s': %v", path, err)
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error: no se pudo leer el MBR del disco '%s': %v", path, err)
	}

	// Buscar partición por nombre
	var targetPartition *Models.Partition
	var partitionIndex int = -1

	for i, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == name {
			targetPartition = &mbr.Partitions[i]
			partitionIndex = i
			break
		}
	}

	if targetPartition == nil {
		return fmt.Errorf("error: no existe la partición '%s' en el disco '%s'", name, path)
	}

	// Validar que es partición primaria
	if targetPartition.PartType != 'P' {
		return fmt.Errorf("error: solo se pueden montar particiones primarias. La partición '%s' es de tipo '%c'", name, targetPartition.PartType)
	}

	// Determinar letra del disco
	var diskLetter rune
	if letter, exists := diskLetterMap[path]; exists {
		// Disco ya tiene letra asignada
		diskLetter = letter
	} else {
		// Nuevo disco, asignar siguiente letra disponible
		diskLetter = nextAvailableLetter
		diskLetterMap[path] = diskLetter
		diskPartitionCount[path] = 0 // Inicializar contador
		nextAvailableLetter++
	}

	// Incrementar contador de particiones para este disco
	diskPartitionCount[path]++
	partitionNumber := diskPartitionCount[path]

	// Generar ID: 68 (carnet) + número + letra
	mountID := fmt.Sprintf("68%d%c", partitionNumber, diskLetter)

	// Actualizar partición en el disco
	mbr.Partitions[partitionIndex].PartStatus = 1 // Marcar como montada
	mbr.Partitions[partitionIndex].PartCorrelative = int64(partitionNumber)

	// Almacenar ID generado en el campo PartID
	copy(mbr.Partitions[partitionIndex].PartID[:], mountID)

	// Escribir MBR actualizado
	file.Seek(0, 0)
	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		return fmt.Errorf("error: no se pudo actualizar el MBR del disco '%s': %v", path, err)
	}

	// Agregar a tabla de montajes en RAM
	mountInfo := MountInfo{
		DiskPath:      path,
		PartitionName: name,
		MountID:       mountID,
		DiskLetter:    diskLetter,
		PartNumber:    partitionNumber,
	}
	mountedPartitions = append(mountedPartitions, mountInfo)

	// Mensaje de éxito
	fmt.Printf("✓ Partición montada exitosamente\n")
	fmt.Printf("  ID: %s\n", mountID)
	fmt.Printf("  Partición: %s\n", name)
	fmt.Printf("  Disco: %s (Letra: %c)\n", path, diskLetter)
	fmt.Printf("  Número de partición en disco: %d\n", partitionNumber)

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

// GetMountedPartitions devuelve la lista de particiones montadas
func GetMountedPartitions() []MountInfo {
	return mountedPartitions
}

// UnmountPartition desmonta una partición (función auxiliar para futuros comandos)
func UnmountPartition(mountID string) error {
	initMountSystem()

	for i, mount := range mountedPartitions {
		if mount.MountID == mountID {
			// Remover de tabla de montajes
			mountedPartitions = append(mountedPartitions[:i], mountedPartitions[i+1:]...)

			// Actualizar contador de particiones del disco
			diskPartitionCount[mount.DiskPath]--

			// Si no quedan particiones montadas del disco, liberar letra
			if diskPartitionCount[mount.DiskPath] == 0 {
				delete(diskLetterMap, mount.DiskPath)
				delete(diskPartitionCount, mount.DiskPath)
			}

			fmt.Printf("✓ Partición %s desmontada exitosamente\n", mountID)
			return nil
		}
	}

	return fmt.Errorf("error: no existe partición montada con ID '%s'", mountID)
}

// ShowMountedPartitions muestra todas las particiones montadas
func ShowMountedPartitions() {
	initMountSystem()

	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas actualmente")
		return
	}

	fmt.Println("=== PARTICIONES MONTADAS ===")
	for _, mount := range mountedPartitions {
		fmt.Printf("ID: %s | Partición: %s | Disco: %s | Letra: %c | Número: %d\n",
			mount.MountID, mount.PartitionName, mount.DiskPath, mount.DiskLetter, mount.PartNumber)
	}
}
