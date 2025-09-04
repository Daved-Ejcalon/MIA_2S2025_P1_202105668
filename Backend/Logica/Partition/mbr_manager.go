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

// MBRManager maneja operaciones sobre el Master Boot Record del disco
type MBRManager struct {
	diskPath string
}

func NewMBRManager(diskPath string) *MBRManager {
	return &MBRManager{
		diskPath: diskPath,
	}
}

// CreateMBR crea un MBR nuevo con tabla de particiones inicializada
func (m *MBRManager) CreateMBR(diskSize int64, fitType byte) (*Models.MBR, error) {
	if !Models.IsValidFitType(fitType) {
		return nil, fmt.Errorf("algoritmo de ajuste invalido")
	}

	if diskSize <= Models.MBR_SIZE {
		return nil, fmt.Errorf("disco demasiado pequeno para MBR")
	}

	mbr := &Models.MBR{
		MbrSize:         diskSize,
		MbrCreationDate: time.Now().Unix(),
		MbrSignature:    rand.Int63(),
		DiskFit:         fitType,
	}

	// Inicializar tabla de 4 particiones vacías
	for i := 0; i < 4; i++ {
		mbr.Partitions[i] = Models.Partition{
			PartStatus:      Models.PARTITION_INACTIVE,
			PartType:        0,
			PartFit:         fitType,
			PartStart:       0,
			PartSize:        0,
			PartCorrelative: -1,
		}

		for j := range mbr.Partitions[i].PartName {
			mbr.Partitions[i].PartName[j] = 0
		}
		for j := range mbr.Partitions[i].PartID {
			mbr.Partitions[i].PartID[j] = 0
		}
	}

	return mbr, nil
}

func (m *MBRManager) WriteMBR(mbr *Models.MBR) error {
	file, err := os.OpenFile(m.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error posicionandose en el archivo")
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, mbr)
	if err != nil {
		return fmt.Errorf("error serializando MBR")
	}

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	return nil
}

func (m *MBRManager) ReadMBR() (*Models.MBR, error) {
	if _, err := os.Stat(m.diskPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo de disco no existe: %s", m.diskPath)
	}

	file, err := os.Open(m.diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error posicionándose")
	}

	mbrBytes := make([]byte, Models.MBR_SIZE)
	_, err = file.Read(mbrBytes)
	if err != nil {
		return nil, fmt.Errorf("error leyendo MBR")
	}

	mbr := &Models.MBR{}
	buffer := bytes.NewReader(mbrBytes)
	err = binary.Read(buffer, binary.LittleEndian, mbr)
	if err != nil {
		return nil, fmt.Errorf("error deserializando MBR")
	}

	return mbr, nil
}

// AddPartition agrega una nueva partición al MBR usando algoritmos de ajuste
func (m *MBRManager) AddPartition(name string, size int64, partType byte, fitType byte) error {
	mbr, err := m.ReadMBR()
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	if !Models.IsValidPartitionType(partType) {
		return fmt.Errorf("tipo de particion invalido: %c. Use P o E", partType)
	}

	if !Models.IsValidFitType(fitType) {
		return fmt.Errorf("tipo de ajuste invalido: %c. Use B, F o W", fitType)
	}

	if size <= 0 {
		return errors.New("el tamano de la particion debe ser mayor a 0")
	}

	// Validar límite de una partición extendida por disco
	if partType == Models.PARTITION_EXTENDED {
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].IsExtended() && !mbr.Partitions[i].IsEmptyPartition() {
				return errors.New("solo puede existir una particion extendida por disco")
			}
		}
	}

	partitionIndex, startPos, err := m.findAvailableSpace(mbr, size, fitType)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar espacio: %v", err)
	}

	partition := &mbr.Partitions[partitionIndex]
	partition.PartStatus = Models.PARTITION_INACTIVE
	partition.PartType = partType
	partition.PartFit = fitType
	partition.PartStart = startPos
	partition.PartSize = size
	partition.PartCorrelative = -1
	partition.SetPartitionName(name)

	err = m.WriteMBR(mbr)
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	return nil
}

// findAvailableSpace busca espacio disponible usando algoritmos FF/BF/WF
func (m *MBRManager) findAvailableSpace(mbr *Models.MBR, size int64, fitType byte) (int, int64, error) {
	// Buscar slot disponible en tabla de particiones
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

	// Mapear espacios ocupados incluyendo MBR y particiones existentes
	occupiedSpaces := make([][2]int64, 0)

	occupiedSpaces = append(occupiedSpaces, [2]int64{0, Models.MBR_SIZE})

	for i := 0; i < 4; i++ {
		partition := &mbr.Partitions[i]
		if !partition.IsEmptyPartition() {
			occupiedSpaces = append(occupiedSpaces, [2]int64{
				partition.PartStart,
				partition.GetPartitionEnd(),
			})
		}
	}

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

// findFirstFit implementa algoritmo First Fit - primer espacio que ajuste
func (m *MBRManager) findFirstFit(occupied [][2]int64, size int64, diskSize int64) int64 {
	// Ordenar espacios ocupados por posición de inicio
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSize := availableEnd - availableStart

		if availableSize >= size {
			return availableStart
		}
	}

	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		if diskSize-lastEnd >= size {
			return lastEnd
		}
	}

	return -1
}

// findBestFit implementa algoritmo Best Fit - menor espacio que ajuste
func (m *MBRManager) findBestFit(occupied [][2]int64, size int64, diskSize int64) int64 {
	bestStart := int64(-1)
	bestSize := int64(diskSize)

	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSize := availableEnd - availableStart

		if availableSize >= size && availableSize < bestSize {
			bestStart = availableStart
			bestSize = availableSize
		}
	}

	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := diskSize - lastEnd
		if finalSpace >= size && finalSpace < bestSize {
			bestStart = lastEnd
		}
	}

	return bestStart
}

// findWorstFit implementa algoritmo Worst Fit - mayor espacio disponible
func (m *MBRManager) findWorstFit(occupied [][2]int64, size int64, diskSize int64) int64 {
	worstStart := int64(-1)
	worstSize := int64(-1)

	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSize := availableEnd - availableStart

		if availableSize >= size && availableSize > worstSize {
			worstStart = availableStart
			worstSize = availableSize
		}
	}

	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := diskSize - lastEnd
		if finalSpace >= size && finalSpace > worstSize {
			worstStart = lastEnd
		}
	}

	return worstStart
}

// RemovePartition elimina una partición del MBR por nombre
func (m *MBRManager) RemovePartition(partitionName string) error {
	mbr, err := m.ReadMBR()
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Buscar partición por nombre
	partitionIndex := -1
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].GetPartitionName() == partitionName {
			partitionIndex = i
			break
		}
	}

	if partitionIndex == -1 {
		return fmt.Errorf("particion '%s' no encontrada", partitionName)
	}

	partition := &mbr.Partitions[partitionIndex]
	*partition = Models.Partition{
		PartStatus:      Models.PARTITION_INACTIVE,
		PartCorrelative: -1,
	}

	err = m.WriteMBR(mbr)
	if err != nil {
		return fmt.Errorf("error escribiendo MBR")
	}

	return nil
}

// GetPartitions retorna lista de particiones activas del MBR
func (m *MBRManager) GetPartitions() ([]Models.Partition, error) {
	mbr, err := m.ReadMBR()
	if err != nil {
		return nil, fmt.Errorf("error leyendo MBR")
	}

	// Recopilar solo particiones no vacías
	partitions := make([]Models.Partition, 0)
	for i := 0; i < 4; i++ {
		if !mbr.Partitions[i].IsEmptyPartition() {
			partitions = append(partitions, mbr.Partitions[i])
		}
	}

	return partitions, nil
}

// ValidateMBR verifica integridad del MBR y sus particiones
func (m *MBRManager) ValidateMBR() error {
	mbr, err := m.ReadMBR()
	if err != nil {
		return fmt.Errorf("error leyendo MBR")
	}

	// Validar metadatos básicos del MBR
	if mbr.MbrSignature == 0 {
		return errors.New("firma del disco invalida")
	}

	if mbr.MbrSize <= Models.MBR_SIZE {
		return errors.New("tamano del disco invalido")
	}

	if !Models.IsValidFitType(mbr.DiskFit) {
		return errors.New("tipo de ajuste del disco invalido")
	}

	// Validar límites de particiones y restricción de extendidas
	extendedCount := 0
	for i := 0; i < 4; i++ {
		partition := &mbr.Partitions[i]

		if !partition.IsEmptyPartition() {
			if partition.PartStart < Models.MBR_SIZE {
				return fmt.Errorf("particion %d inicia antes del final del MBR", i)
			}

			if partition.GetPartitionEnd() > mbr.MbrSize {
				return fmt.Errorf("particion %d excede el tamano del disco", i)
			}

			if partition.IsExtended() {
				extendedCount++
			}
		}
	}

	if extendedCount > 1 {
		return errors.New("no puede haber mas de una particion extendida")
	}

	return nil
}
