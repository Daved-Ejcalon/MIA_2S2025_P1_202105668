package Disk

import (
	"MIA_2S2025_P1_202105668/Logica/Partition"
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
)

// MountInfo muestra la información de una partición montada
type MountInfo struct {
	DiskPath      string // Ruta
	PartitionName string // Nombre de la partición
	MountID       string // ID ("681A")
	DiskLetter    rune   // Letra
	PartNumber    int    // Número de partición
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
		return fmt.Errorf("path requerido")
	}
	if name == "" {
		return fmt.Errorf("nombre requerido")
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("archivo no existe")
	}

	// Verificar si ya está montada
	if isAlreadyMounted(path, name) {
		return fmt.Errorf("partición ya montada")
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

	// Buscar partición por nombre
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

	// Validar que es partición primaria únicamente
	if targetPartition.PartType != 'P' {
		return fmt.Errorf("solo se pueden montar particiones primarias")
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

	// Generar ID: 68 (últimos dos dígitos del carnet 202105668) + número + letra
	mountID := fmt.Sprintf("68%d%c", partitionNumber, diskLetter)

	// Actualizar atributos de la partición EN MEMORIA (no escribir al disco)
	targetPartition.PartStatus = 1 // Marcar como montada
	targetPartition.PartCorrelative = int64(partitionNumber)
	
	// Almacenar ID generado en el campo PartID (solo en memoria)
	copy(targetPartition.PartID[:], mountID)

	// Agregar a tabla de montajes en RAM
	mountInfo := MountInfo{
		DiskPath:      path,
		PartitionName: name,
		MountID:       mountID,
		DiskLetter:    diskLetter,
		PartNumber:    partitionNumber,
	}
	mountedPartitions = append(mountedPartitions, mountInfo)

	fmt.Printf("Partición montada: %s\n", mountID)

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

			fmt.Printf("Partición desmontada: %s\n", mountID)
			return nil
		}
	}

	return fmt.Errorf("ID no encontrado")
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

// mountLogicalPartition monta una partición lógica buscándola en la cadena de EBRs
func mountLogicalPartition(path string, name string, logicalPartition *Models.Partition) error {
	// Buscar la partición lógica en la cadena de EBRs
	ebrMgr := Partition.NewEBRManager(path, logicalPartition)

	// Verificar que la partición lógica existe
	exists, err := ebrMgr.LogicalPartitionExists(name)
	if err != nil {
		return fmt.Errorf("error verificando partición")
	}
	if !exists {
		return fmt.Errorf("partición lógica no encontrada")
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

	// Actualizar partición en el disco (estado de montaje)
	logicalPartition.PartStatus = 1 // Marcar como montada
	logicalPartition.PartCorrelative = int64(partitionNumber)
	logicalPartition.SetPartitionID(mountID)

	// Agregar a tabla de montajes en RAM
	mountInfo := MountInfo{
		DiskPath:      path,
		PartitionName: name,
		MountID:       mountID,
		DiskLetter:    diskLetter,
		PartNumber:    partitionNumber,
	}
	mountedPartitions = append(mountedPartitions, mountInfo)

	fmt.Printf("Partición lógica montada: %s\n", mountID)

	return nil
}
