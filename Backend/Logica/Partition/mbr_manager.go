// Package Partition maneja las operaciones de particiones y MBR
package Partition

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// MBRManager maneja todas las operaciones del Master Boot Record
type MBRManager struct {
	diskPath string // Ruta del archivo de disco
}

// NewMBRManager crea una nueva instancia del manager de MBR
func NewMBRManager(diskPath string) *MBRManager {
	return &MBRManager{
		diskPath: diskPath,
	}
}

// CreateMBR crea un nuevo MBR con los parámetros especificados
func (m *MBRManager) CreateMBR(diskSize int64, fitType byte) (*Models.MBR, error) {
	// Validar tipo de ajuste
	if !Models.IsValidFitType(fitType) {
		return nil, fmt.Errorf("fit inválido")
	}

	// Validar tamaño del disco
	if diskSize <= Models.MBR_SIZE {
		return nil, fmt.Errorf("tamaño inválido")
	}

	// Crear nuevo MBR
	mbr := &Models.MBR{
		MbrSize:         diskSize,
		MbrCreationDate: time.Now().Unix(),
		MbrSignature:    rand.Int63(), // Número random para identificar el disco
		DiskFit:         fitType,
	}

	// Inicializar particiones vacías
	for i := 0; i < 4; i++ {
		mbr.Partitions[i] = Models.Partition{
			PartStatus:      Models.PARTITION_INACTIVE,
			PartType:        0,
			PartFit:         fitType, // Heredar el tipo de ajuste del MBR
			PartStart:       0,
			PartSize:        0,
			PartCorrelative: -1, // -1 indica partición no montada
		}

		// Limpiar nombre e ID
		for j := range mbr.Partitions[i].PartName {
			mbr.Partitions[i].PartName[j] = 0
		}
		for j := range mbr.Partitions[i].PartID {
			mbr.Partitions[i].PartID[j] = 0
		}
	}

	return mbr, nil
}

// WriteMBR escribe el MBR al archivo de disco en el offset 0
func (m *MBRManager) WriteMBR(mbr *Models.MBR) error {
	// Abrir archivo en modo lectura/escritura
	file, err := os.OpenFile(m.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	// Posicionarse al inicio del archivo (offset 0)
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error posicionándose")
	}

	// Serializar el MBR a bytes
	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, mbr)
	if err != nil {
		return fmt.Errorf("error serializando MBR")
	}

	// Escribir al archivo
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	return nil
}

// ReadMBR lee el MBR desde el archivo de disco
func (m *MBRManager) ReadMBR() (*Models.MBR, error) {
	// Verificar que el archivo existe
	if _, err := os.Stat(m.diskPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo de disco no existe: %s", m.diskPath)
	}

	// Abrir archivo en modo lectura
	file, err := os.Open(m.diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	// Posicionarse al inicio del archivo
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error posicionándose")
	}

	// Leer bytes del MBR
	mbrBytes := make([]byte, Models.MBR_SIZE)
	_, err = file.Read(mbrBytes)
	if err != nil {
		return nil, fmt.Errorf("error leyendo MBR")
	}

	// Deserializar bytes a estructura MBR
	mbr := &Models.MBR{}
	buffer := bytes.NewReader(mbrBytes)
	err = binary.Read(buffer, binary.LittleEndian, mbr)
	if err != nil {
		return nil, fmt.Errorf("error deserializando MBR")
	}

	return mbr, nil
}

// AddPartition agrega una nueva partición al MBR
func (m *MBRManager) AddPartition(name string, size int64, partType byte, fitType byte) error {
	// Leer MBR actual
	mbr, err := m.ReadMBR()
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Validar parámetros
	if !Models.IsValidPartitionType(partType) {
		return fmt.Errorf("tipo de partición inválido: %c. Use P o E", partType)
	}

	if !Models.IsValidFitType(fitType) {
		return fmt.Errorf("tipo de ajuste inválido: %c. Use B, F o W", fitType)
	}

	if size <= 0 {
		return errors.New("el tamaño de la partición debe ser mayor a 0")
	}

	// Verificar que no existe ya una partición extendida si se está creando una
	if partType == Models.PARTITION_EXTENDED {
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].IsExtended() && !mbr.Partitions[i].IsEmptyPartition() {
				return errors.New("solo puede existir una partición extendida por disco")
			}
		}
	}

	// Buscar espacio disponible según el tipo de ajuste
	partitionIndex, startPos, err := m.findAvailableSpace(mbr, size, fitType)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar espacio: %v", err)
	}

	// Configurar la nueva partición
	partition := &mbr.Partitions[partitionIndex]
	partition.PartStatus = Models.PARTITION_INACTIVE
	partition.PartType = partType
	partition.PartFit = fitType
	partition.PartStart = startPos
	partition.PartSize = size
	partition.PartCorrelative = -1
	partition.SetPartitionName(name)

	// Escribir MBR actualizado
	err = m.WriteMBR(mbr)
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	return nil
}

// findAvailableSpace busca espacio disponible para una nueva partición
func (m *MBRManager) findAvailableSpace(mbr *Models.MBR, size int64, fitType byte) (int, int64, error) {
	// Buscar slot vacío en el MBR
	partitionIndex := -1
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].IsEmptyPartition() {
			partitionIndex = i
			break
		}
	}

	if partitionIndex == -1 {
		return -1, 0, errors.New("no hay slots disponibles en el MBR")
	}

	// Crear lista de espacios ocupados
	occupiedSpaces := make([][2]int64, 0)

	// El MBR ocupa el primer espacio
	occupiedSpaces = append(occupiedSpaces, [2]int64{0, Models.MBR_SIZE})

	// Agregar particiones existentes
	for i := 0; i < 4; i++ {
		partition := &mbr.Partitions[i]
		if !partition.IsEmptyPartition() {
			occupiedSpaces = append(occupiedSpaces, [2]int64{
				partition.PartStart,
				partition.GetPartitionEnd(),
			})
		}
	}

	// Buscar espacio según el tipo de ajuste
	switch fitType {
	case Models.FIT_FIRST:
		return partitionIndex, m.findFirstFit(occupiedSpaces, size, mbr.MbrSize), nil
	case Models.FIT_BEST:
		return partitionIndex, m.findBestFit(occupiedSpaces, size, mbr.MbrSize), nil
	case Models.FIT_WORST:
		return partitionIndex, m.findWorstFit(occupiedSpaces, size, mbr.MbrSize), nil
	default:
		return -1, 0, errors.New("tipo de ajuste no reconocido")
	}
}

// findFirstFit implementa el algoritmo First Fit
func (m *MBRManager) findFirstFit(occupied [][2]int64, size int64, diskSize int64) int64 {
	// Ordenar espacios ocupados por posición inicial
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	// Buscar el primer espacio disponible
	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSize := availableEnd - availableStart

		if availableSize >= size {
			return availableStart
		}
	}

	// Verificar espacio al final del disco
	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		if diskSize-lastEnd >= size {
			return lastEnd
		}
	}

	return -1 // No hay espacio suficiente
}

// findBestFit implementa el algoritmo Best Fit
func (m *MBRManager) findBestFit(occupied [][2]int64, size int64, diskSize int64) int64 {
	bestStart := int64(-1)
	bestSize := int64(diskSize) // Inicializar con el tamaño máximo

	// Ordenar espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	// Buscar entre espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSize := availableEnd - availableStart

		if availableSize >= size && availableSize < bestSize {
			bestStart = availableStart
			bestSize = availableSize
		}
	}

	// Verificar espacio al final
	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := diskSize - lastEnd
		if finalSpace >= size && finalSpace < bestSize {
			bestStart = lastEnd
		}
	}

	return bestStart
}

// findWorstFit implementa el algoritmo Worst Fit
func (m *MBRManager) findWorstFit(occupied [][2]int64, size int64, diskSize int64) int64 {
	worstStart := int64(-1)
	worstSize := int64(-1)

	// Ordenar espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	// Buscar entre espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSize := availableEnd - availableStart

		if availableSize >= size && availableSize > worstSize {
			worstStart = availableStart
			worstSize = availableSize
		}
	}

	// Verificar espacio al final
	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := diskSize - lastEnd
		if finalSpace >= size && finalSpace > worstSize {
			worstStart = lastEnd
		}
	}

	return worstStart
}

// RemovePartition elimina una partición del MBR
func (m *MBRManager) RemovePartition(partitionName string) error {
	mbr, err := m.ReadMBR()
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Buscar la partición por nombre
	partitionIndex := -1
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].GetPartitionName() == partitionName {
			partitionIndex = i
			break
		}
	}

	if partitionIndex == -1 {
		return fmt.Errorf("partición '%s' no encontrada", partitionName)
	}

	// Limpiar la partición
	partition := &mbr.Partitions[partitionIndex]
	*partition = Models.Partition{
		PartStatus:      Models.PARTITION_INACTIVE,
		PartCorrelative: -1,
	}

	// Escribir MBR actualizado
	err = m.WriteMBR(mbr)
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	return nil
}

// GetPartitions retorna la lista de particiones del MBR
func (m *MBRManager) GetPartitions() ([]Models.Partition, error) {
	mbr, err := m.ReadMBR()
	if err != nil {
		return nil, fmt.Errorf("error leyendo MBR")
	}

	partitions := make([]Models.Partition, 0)
	for i := 0; i < 4; i++ {
		if !mbr.Partitions[i].IsEmptyPartition() {
			partitions = append(partitions, mbr.Partitions[i])
		}
	}

	return partitions, nil
}

// ValidateMBR valida la consistencia del MBR
func (m *MBRManager) ValidateMBR() error {
	mbr, err := m.ReadMBR()
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Validar firma del disco
	if mbr.MbrSignature == 0 {
		return errors.New("firma del disco inválida")
	}

	// Validar tamaño del disco
	if mbr.MbrSize <= Models.MBR_SIZE {
		return errors.New("tamaño del disco inválido")
	}

	// Validar tipo de ajuste del disco
	if !Models.IsValidFitType(mbr.DiskFit) {
		return errors.New("tipo de ajuste del disco inválido")
	}

	// Validar particiones
	extendedCount := 0
	for i := 0; i < 4; i++ {
		partition := &mbr.Partitions[i]

		if !partition.IsEmptyPartition() {
			// Validar que la partición esté dentro del disco
			if partition.PartStart < Models.MBR_SIZE {
				return fmt.Errorf("partición %d inicia antes del final del MBR", i)
			}

			if partition.GetPartitionEnd() > mbr.MbrSize {
				return fmt.Errorf("partición %d excede el tamaño del disco", i)
			}

			// Contar particiones extendidas
			if partition.IsExtended() {
				extendedCount++
			}
		}
	}

	// Validar que no haya más de una partición extendida
	if extendedCount > 1 {
		return errors.New("no puede haber más de una partición extendida")
	}

	return nil
}
