package Partition

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// EBRManager maneja particiones lógicas usando Extended Boot Records
type EBRManager struct {
	diskPath          string
	extendedPartition *Models.Partition
}

func NewEBRManager(diskPath string, extendedPartition *Models.Partition) *EBRManager {
	return &EBRManager{
		diskPath:          diskPath,
		extendedPartition: extendedPartition,
	}
}

// CreateFirstEBR inicializa el primer EBR vacío en la partición extendida solo si no existe
func (e *EBRManager) CreateFirstEBR() error {
	if e.extendedPartition == nil {
		return errors.New("se requiere una particion extendida valida")
	}

	if !e.extendedPartition.IsExtended() {
		return errors.New("la particion debe ser de tipo extendida")
	}

	// Verificar si ya existe un EBR válido en el inicio de la partición extendida
	existingEBR, err := e.ReadEBR(e.extendedPartition.PartStart)
	if err == nil {
		// Verificar si el EBR es válido (no vacío y no con partición lógica activa)
		if !existingEBR.IsEmptyEBR() {
			return nil
		}
		// Si es un EBR vacío válido con PartNext correcto, tampoco sobrescribir
		if existingEBR.IsEmptyEBR() && existingEBR.PartNext == Models.EBR_END {
			return nil
		}
	}


	firstEBR := Models.EBR{
		PartMount: Models.EBR_UNMOUNTED,
		PartFit:   e.extendedPartition.PartFit,
		PartStart: 0,
		PartS:     0,
		PartNext:  Models.EBR_END,
	}

	for i := range firstEBR.PartName {
		firstEBR.PartName[i] = 0
	}

	err = e.WriteEBR(&firstEBR, e.extendedPartition.PartStart)
	if err != nil {
		return fmt.Errorf("error creando EBR")
	}

	return nil
}

func (e *EBRManager) WriteEBR(ebr *Models.EBR, position int64) error {
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	_, err = file.Seek(position, 0)
	if err != nil {
		return fmt.Errorf("error posicionandose en el archivo")
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, ebr)
	if err != nil {
		return fmt.Errorf("error serializando EBR")
	}

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("error escribiendo EBR")
	}

	// Forzar sincronización al disco
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("error sincronizando EBR")
	}

	return nil
}

func (e *EBRManager) ReadEBR(position int64) (*Models.EBR, error) {
	if _, err := os.Stat(e.diskPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo de disco no existe: %s", e.diskPath)
	}

	file, err := os.Open(e.diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	_, err = file.Seek(position, 0)
	if err != nil {
		return nil, fmt.Errorf("error posicionándose")
	}

	ebrBytes := make([]byte, Models.EBR_SIZE)
	_, err = file.Read(ebrBytes)
	if err != nil {
		return nil, fmt.Errorf("error leyendo EBR")
	}

	ebr := &Models.EBR{}
	buffer := bytes.NewReader(ebrBytes)
	err = binary.Read(buffer, binary.LittleEndian, ebr)
	if err != nil {
		return nil, fmt.Errorf("error deserializando EBR")
	}


	return ebr, nil
}

// AddLogicalPartition crea una nueva partición lógica usando algoritmos de ajuste
func (e *EBRManager) AddLogicalPartition(name string, size int64, fitType byte) error {

	// Validaciones de entrada
	if name == "" {
		return errors.New("el nombre de la partición lógica es obligatorio")
	}

	if size <= 0 {
		return errors.New("el tamaño de la partición lógica debe ser mayor a 0")
	}

	if !Models.IsValidFitType(fitType) {
		return fmt.Errorf("tipo de ajuste inválido: %c. Use B, F o W", fitType)
	}

	exists, err := e.LogicalPartitionExists(name)
	if err != nil {
		return fmt.Errorf("error verificando particiones")
	}
	if exists {
		return fmt.Errorf("ya existe una partición lógica con el nombre '%s'", name)
	}

	insertPosition, err := e.findInsertPosition(size, fitType)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar espacio para la partición: %v", err)
	}


	currentEBR, ebrPosition, err := e.getEBRForInsertion(insertPosition)
	if err != nil {
		return fmt.Errorf("error obteniendo EBR")
	}


	if currentEBR.IsEmptyEBR() {
		currentEBR.PartMount = Models.EBR_UNMOUNTED
		currentEBR.PartFit = fitType
		currentEBR.PartStart = insertPosition + Models.EBR_SIZE
		currentEBR.PartS = size
		currentEBR.SetLogicalPartitionName(name)
	} else {
		return e.insertNewEBRInChain(name, size, fitType, insertPosition)
	}


	err = e.WriteEBR(currentEBR, ebrPosition)
	if err != nil {
		return fmt.Errorf("error escribiendo EBR")
	}


	return nil
}

// findInsertPosition encuentra posición según algoritmo de ajuste (FF/BF/WF)
func (e *EBRManager) findInsertPosition(size int64, fitType byte) (int64, error) {
	occupiedSpaces, err := e.getOccupiedSpacesInExtended()
	if err != nil {
		return -1, fmt.Errorf("error obteniendo espacios")
	}

	extendedStart := e.extendedPartition.PartStart
	extendedEnd := e.extendedPartition.GetPartitionEnd()

	// Aplicar algoritmo de ajuste específico
	switch fitType {
	case Models.FIT_FIRST:
		return e.findFirstFitInExtended(occupiedSpaces, size, extendedStart, extendedEnd), nil
	case Models.FIT_BEST:
		return e.findBestFitInExtended(occupiedSpaces, size, extendedStart, extendedEnd), nil
	case Models.FIT_WORST:
		return e.findWorstFitInExtended(occupiedSpaces, size, extendedStart, extendedEnd), nil
	default:
		return -1, errors.New("tipo de ajuste no reconocido")
	}
}

// getOccupiedSpacesInExtended recorre cadena EBR para mapear espacios ocupados
func (e *EBRManager) getOccupiedSpacesInExtended() ([][2]int64, error) {
	occupiedSpaces := make([][2]int64, 0)

	// Recorrer cadena de EBRs desde el inicio
	currentEBRPos := e.extendedPartition.PartStart

	for currentEBRPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentEBRPos)
		if err != nil {
			return nil, fmt.Errorf("error leyendo EBR")
		}


		if !ebr.IsEmptyEBR() {
			spaceRange := [2]int64{currentEBRPos, ebr.GetPartitionEnd()}
			occupiedSpaces = append(occupiedSpaces, spaceRange)
		}

		if ebr.HasNext() {
			currentEBRPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return occupiedSpaces, nil
}

// findFirstFitInExtended implementa algoritmo First Fit para encontrar espacio
func (e *EBRManager) findFirstFitInExtended(occupied [][2]int64, size int64, extStart int64, extEnd int64) int64 {
	// Ordenar espacios ocupados por posición de inicio
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	if len(occupied) == 0 || occupied[0][0] > extStart+Models.EBR_SIZE+size {
		return extStart
	}

	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]

		if availableEnd-availableStart >= Models.EBR_SIZE+size {
			return availableStart
		}
	}

	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		if extEnd-lastEnd >= Models.EBR_SIZE+size {
			return lastEnd
		}
	}

	return -1
}

// findBestFitInExtended implementa algoritmo Best Fit (menor espacio que ajuste)
func (e *EBRManager) findBestFitInExtended(occupied [][2]int64, size int64, extStart int64, extEnd int64) int64 {
	bestStart := int64(-1)
	bestSize := extEnd - extStart
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	if len(occupied) == 0 {
		return extStart
	}

	if occupied[0][0] > extStart {
		spaceAtStart := occupied[0][0] - extStart
		if spaceAtStart >= Models.EBR_SIZE+size && spaceAtStart < bestSize {
			bestStart = extStart
			bestSize = spaceAtStart
		}
	}

	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSpace := availableEnd - availableStart

		if availableSpace >= Models.EBR_SIZE+size && availableSpace < bestSize {
			bestStart = availableStart
			bestSize = availableSpace
		}
	}

	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := extEnd - lastEnd
		if finalSpace >= Models.EBR_SIZE+size && finalSpace < bestSize {
			bestStart = lastEnd
		}
	}

	return bestStart
}

// findWorstFitInExtended implementa algoritmo Worst Fit (mayor espacio disponible)
func (e *EBRManager) findWorstFitInExtended(occupied [][2]int64, size int64, extStart int64, extEnd int64) int64 {
	worstStart := int64(-1)
	worstSize := int64(-1)

	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	if len(occupied) == 0 {
		return extStart
	}

	if occupied[0][0] > extStart {
		spaceAtStart := occupied[0][0] - extStart
		if spaceAtStart >= Models.EBR_SIZE+size && spaceAtStart > worstSize {
			worstStart = extStart
			worstSize = spaceAtStart
		}
	}

	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSpace := availableEnd - availableStart

		if availableSpace >= Models.EBR_SIZE+size && availableSpace > worstSize {
			worstStart = availableStart
			worstSize = availableSpace
		}
	}

	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := extEnd - lastEnd
		if finalSpace >= Models.EBR_SIZE+size && finalSpace > worstSize {
			worstStart = lastEnd
		}
	}

	return worstStart
}

func (e *EBRManager) getEBRForInsertion(position int64) (*Models.EBR, int64, error) {
	if position == e.extendedPartition.PartStart {
		ebr, err := e.ReadEBR(position)
		return ebr, position, err
	}
	currentPos := e.extendedPartition.PartStart
	var lastEBRPos int64 = currentPos

	for currentPos != Models.EBR_END && currentPos < position {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return nil, -1, fmt.Errorf("error leyendo EBR")
		}

		lastEBRPos = currentPos

		if ebr.HasNext() {
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	ebr, err := e.ReadEBR(lastEBRPos)
	return ebr, lastEBRPos, err
}

func (e *EBRManager) insertNewEBRInChain(name string, size int64, fitType byte, position int64) error {
	newEBR := Models.EBR{
		PartMount: Models.EBR_UNMOUNTED,
		PartFit:   fitType,
		PartStart: position + Models.EBR_SIZE,
		PartS:     size,
		PartNext:  Models.EBR_END,
	}
	newEBR.SetLogicalPartitionName(name)
	err := e.updateEBRChainForInsertion(&newEBR, position)
	if err != nil {
		return fmt.Errorf("error actualizando cadena EBRs")
	}

	err = e.WriteEBR(&newEBR, position)
	if err != nil {
		return fmt.Errorf("error escribiendo EBR")
	}

	return nil
}

// updateEBRChainForInsertion actualiza punteros de la lista enlazada de EBRs
func (e *EBRManager) updateEBRChainForInsertion(newEBR *Models.EBR, newPosition int64) error {
	currentPos := e.extendedPartition.PartStart

	// Recorrer cadena para encontrar punto de inserción
	for currentPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR")
		}

		if ebr.HasNext() && ebr.GetNextEBRPosition() > newPosition {
			nextEBRPos := ebr.GetNextEBRPosition()
			ebr.SetNextEBRPosition(newPosition)
			newEBR.SetNextEBRPosition(nextEBRPos)

			err = e.WriteEBR(ebr, currentPos)
			if err != nil {
				return fmt.Errorf("error actualizando EBR")
			}

			return nil
		}

		if !ebr.HasNext() {
			ebr.SetNextEBRPosition(newPosition)
			newEBR.MarkAsLastEBR()

			err = e.WriteEBR(ebr, currentPos)
			if err != nil {
				return fmt.Errorf("error actualizando EBR")
			}

			return nil
		}

		currentPos = ebr.GetNextEBRPosition()
	}

	return errors.New("no se pudo encontrar posición de inserción en la cadena")
}

// LogicalPartitionExists verifica si existe una partición lógica con el nombre dado
func (e *EBRManager) LogicalPartitionExists(name string) (bool, error) {
	logicalPartitions, err := e.GetLogicalPartitions()
	if err != nil {
		return false, fmt.Errorf("error obteniendo particiones")
	}

	// Buscar nombre en lista de particiones lógicas
	for _, partition := range logicalPartitions {
		if partition.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetLogicalPartitions retorna lista de todas las particiones lógicas activas
func (e *EBRManager) GetLogicalPartitions() ([]Models.LogicalPartitionInfo, error) {
	logicalPartitions := make([]Models.LogicalPartitionInfo, 0)
	currentPos := e.extendedPartition.PartStart

	// Recorrer cadena de EBRs recolectando particiones activas
	for currentPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return nil, fmt.Errorf("error leyendo EBR")
		}

		if !ebr.IsEmptyEBR() {
			partitionInfo := ebr.ToLogicalPartitionInfo(currentPos)
			logicalPartitions = append(logicalPartitions, partitionInfo)
		}

		if ebr.HasNext() {
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return logicalPartitions, nil
}

func (e *EBRManager) RemoveLogicalPartition(partitionName string) error {
	currentPos := e.extendedPartition.PartStart
	var previousPos int64 = -1

	for currentPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR")
		}

		if !ebr.IsEmptyEBR() && ebr.GetLogicalPartitionName() == partitionName {
			return e.removeEBRFromChain(currentPos, previousPos, ebr)
		}

		if ebr.HasNext() {
			previousPos = currentPos
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return fmt.Errorf("partición lógica '%s' no encontrada", partitionName)
}

// removeEBRFromChain elimina EBR de la lista enlazada actualizando punteros
func (e *EBRManager) removeEBRFromChain(ebrPos int64, previousEBRPos int64, ebrToRemove *Models.EBR) error {
	// Si es el primer EBR, solo limpiarlo
	if previousEBRPos == -1 {
		ebrToRemove.ClearEBR()
		return e.WriteEBR(ebrToRemove, ebrPos)
	}

	previousEBR, err := e.ReadEBR(previousEBRPos)
	if err != nil {
		return fmt.Errorf("error leyendo EBR anterior")
	}

	if ebrToRemove.HasNext() {
		previousEBR.SetNextEBRPosition(ebrToRemove.GetNextEBRPosition())
	} else {
		previousEBR.MarkAsLastEBR()
	}

	// Escribir el EBR anterior actualizado
	err = e.WriteEBR(previousEBR, previousEBRPos)
	if err != nil {
		return fmt.Errorf("error actualizando EBR anterior")
	}

	return nil
}

// ValidateEBRChain verifica integridad de la cadena EBR (loops, límites, etc.)
func (e *EBRManager) ValidateEBRChain() error {
	currentPos := e.extendedPartition.PartStart
	visitedPositions := make(map[int64]bool)

	// Recorrer cadena detectando loops y validando límites
	for currentPos != Models.EBR_END {
		if visitedPositions[currentPos] {
			return fmt.Errorf("loop detectado en la cadena de EBRs en posición %d", currentPos)
		}
		visitedPositions[currentPos] = true

		if currentPos < e.extendedPartition.PartStart ||
			currentPos >= e.extendedPartition.GetPartitionEnd() {
			return fmt.Errorf("EBR en posición %d está fuera de la partición extendida", currentPos)
		}

		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR")
		}

		if !ebr.IsEmptyEBR() {
			err = ebr.ValidateEBR()
			if err != nil {
				return fmt.Errorf("EBR inválido en posición %d: %v", currentPos, err)
			}

			if ebr.GetPartitionEnd() > e.extendedPartition.GetPartitionEnd() {
				return fmt.Errorf("partición lógica en EBR %d excede la partición extendida", currentPos)
			}
		}

		if ebr.HasNext() {
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return nil
}
