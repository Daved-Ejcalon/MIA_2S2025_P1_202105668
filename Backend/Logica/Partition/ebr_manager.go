// Package Partition maneja las operaciones de EBR y particiones lógicas
package Partition

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// EBRManager maneja todas las operaciones del Extended Boot Record
type EBRManager struct {
	diskPath          string            // Ruta del archivo de disco
	extendedPartition *Models.Partition // Referencia a la partición extendida
}

// NewEBRManager crea una nueva instancia del manager de EBR
func NewEBRManager(diskPath string, extendedPartition *Models.Partition) *EBRManager {
	return &EBRManager{
		diskPath:          diskPath,
		extendedPartition: extendedPartition,
	}
}

// CreateFirstEBR crea el primer EBR al crear la partición extendida
func (e *EBRManager) CreateFirstEBR() error {
	// Validar que tenemos una partición extendida
	if e.extendedPartition == nil {
		return errors.New("partición extendida requerida")
	}

	if !e.extendedPartition.IsExtended() {
		return errors.New("partición no es extendida")
	}

	// Crear EBR vacío inicial
	firstEBR := Models.EBR{
		PartMount: Models.EBR_UNMOUNTED,
		PartFit:   e.extendedPartition.PartFit, // Heredar fit de la partición extendida
		PartStart: 0,                           // Se establecerá cuando se agregue la primera partición lógica
		PartS:     0,                           // Se establecerá cuando se agregue la primera partición lógica
		PartNext:  Models.EBR_END,              // Inicialmente no hay siguiente
	}

	// Limpiar nombre
	for i := range firstEBR.PartName {
		firstEBR.PartName[i] = 0
	}

	// Escribir el primer EBR al inicio de la partición extendida
	err := e.WriteEBR(&firstEBR, e.extendedPartition.PartStart)
	if err != nil {
		return fmt.Errorf("error creando EBR")
	}

	return nil
}

// WriteEBR escribe un EBR en la posición especificada del disco
func (e *EBRManager) WriteEBR(ebr *Models.EBR, position int64) error {
	// Abrir archivo en modo lectura/escritura
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	// Posicionarse en el offset especificado
	_, err = file.Seek(position, 0)
	if err != nil {
		return fmt.Errorf("error posicionándose")
	}

	// Serializar el EBR a bytes
	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, ebr)
	if err != nil {
		return fmt.Errorf("error serializando EBR")
	}

	// Escribir al archivo
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("error escribiendo EBR")
	}

	return nil
}

// ReadEBR lee un EBR desde la posición especificada del disco
func (e *EBRManager) ReadEBR(position int64) (*Models.EBR, error) {
	// Verificar que el archivo existe
	if _, err := os.Stat(e.diskPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo de disco no existe: %s", e.diskPath)
	}

	// Abrir archivo en modo lectura
	file, err := os.Open(e.diskPath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo disco")
	}
	defer file.Close()

	// Posicionarse en el offset especificado
	_, err = file.Seek(position, 0)
	if err != nil {
		return nil, fmt.Errorf("error posicionándose")
	}

	// Leer bytes del EBR
	ebrBytes := make([]byte, Models.EBR_SIZE)
	_, err = file.Read(ebrBytes)
	if err != nil {
		return nil, fmt.Errorf("error leyendo EBR")
	}

	// Deserializar bytes a estructura EBR
	ebr := &Models.EBR{}
	buffer := bytes.NewReader(ebrBytes)
	err = binary.Read(buffer, binary.LittleEndian, ebr)
	if err != nil {
		return nil, fmt.Errorf("error deserializando EBR")
	}

	return ebr, nil
}

// AddLogicalPartition agrega una nueva partición lógica a la lista enlazada de EBRs
func (e *EBRManager) AddLogicalPartition(name string, size int64, fitType byte) error {
	// Validar parámetros
	if name == "" {
		return errors.New("el nombre de la partición lógica es obligatorio")
	}

	if size <= 0 {
		return errors.New("el tamaño de la partición lógica debe ser mayor a 0")
	}

	if !Models.IsValidFitType(fitType) {
		return fmt.Errorf("tipo de ajuste inválido: %c. Use B, F o W", fitType)
	}

	// Verificar que no existe una partición lógica con el mismo nombre
	exists, err := e.LogicalPartitionExists(name)
	if err != nil {
		return fmt.Errorf("error verificando particiones")
	}
	if exists {
		return fmt.Errorf("ya existe una partición lógica con el nombre '%s'", name)
	}

	// Encontrar posición donde insertar la nueva partición lógica
	insertPosition, err := e.findInsertPosition(size, fitType)
	if err != nil {
		return fmt.Errorf("no se pudo encontrar espacio para la partición: %v", err)
	}

	// Obtener el EBR actual en esa posición o crear uno nuevo
	currentEBR, ebrPosition, err := e.getEBRForInsertion(insertPosition)
	if err != nil {
		return fmt.Errorf("error obteniendo EBR")
	}

	// Configurar la nueva partición lógica
	if currentEBR.IsEmptyEBR() {
		// EBR vacío, configurar con la nueva partición
		currentEBR.PartMount = Models.EBR_UNMOUNTED
		currentEBR.PartFit = fitType
		currentEBR.PartStart = insertPosition + Models.EBR_SIZE // Datos después del EBR
		currentEBR.PartS = size
		currentEBR.SetLogicalPartitionName(name)
		// PartNext se mantiene igual (puede apuntar al siguiente o ser -1)
	} else {
		// Necesitamos insertar un nuevo EBR en la cadena
		return e.insertNewEBRInChain(name, size, fitType, insertPosition)
	}

	// Escribir el EBR actualizado
	err = e.WriteEBR(currentEBR, ebrPosition)
	if err != nil {
		return fmt.Errorf("error escribiendo EBR")
	}

	return nil
}

// findInsertPosition encuentra la posición donde insertar una nueva partición lógica
func (e *EBRManager) findInsertPosition(size int64, fitType byte) (int64, error) {
	// Obtener lista de espacios ocupados dentro de la partición extendida
	occupiedSpaces, err := e.getOccupiedSpacesInExtended()
	if err != nil {
		return -1, fmt.Errorf("error obteniendo espacios")
	}

	extendedStart := e.extendedPartition.PartStart
	extendedEnd := e.extendedPartition.GetPartitionEnd()

	// Buscar espacio según el tipo de ajuste
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

// getOccupiedSpacesInExtended obtiene lista de espacios ocupados dentro de la partición extendida
func (e *EBRManager) getOccupiedSpacesInExtended() ([][2]int64, error) {
	occupiedSpaces := make([][2]int64, 0)

	// Recorrer la lista enlazada de EBRs
	currentEBRPos := e.extendedPartition.PartStart

	for currentEBRPos != Models.EBR_END {
		// Leer EBR actual
		ebr, err := e.ReadEBR(currentEBRPos)
		if err != nil {
			return nil, fmt.Errorf("error leyendo EBR")
		}

		// Si el EBR no está vacío, agregar su espacio ocupado
		if !ebr.IsEmptyEBR() {
			// El EBR ocupa espacio desde su posición hasta el final de los datos
			occupiedSpaces = append(occupiedSpaces, [2]int64{
				currentEBRPos,         // Inicio: posición del EBR
				ebr.GetPartitionEnd(), // Fin: final de los datos de la partición
			})
		}

		// Avanzar al siguiente EBR
		if ebr.HasNext() {
			currentEBRPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return occupiedSpaces, nil
}

// findFirstFitInExtended implementa First Fit dentro de la partición extendida
func (e *EBRManager) findFirstFitInExtended(occupied [][2]int64, size int64, extStart int64, extEnd int64) int64 {
	// Ordenar espacios ocupados por posición
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	// Verificar espacio al inicio de la partición extendida
	if len(occupied) == 0 || occupied[0][0] > extStart+Models.EBR_SIZE+size {
		return extStart
	}

	// Buscar huecos entre espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]

		// Necesitamos espacio para EBR + datos
		if availableEnd-availableStart >= Models.EBR_SIZE+size {
			return availableStart
		}
	}

	// Verificar espacio al final
	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		if extEnd-lastEnd >= Models.EBR_SIZE+size {
			return lastEnd
		}
	}

	return -1 // No hay espacio suficiente
}

// findBestFitInExtended implementa Best Fit dentro de la partición extendida
func (e *EBRManager) findBestFitInExtended(occupied [][2]int64, size int64, extStart int64, extEnd int64) int64 {
	bestStart := int64(-1)
	bestSize := extEnd - extStart // Inicializar con el tamaño máximo

	// Ordenar espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		for j := i + 1; j < len(occupied); j++ {
			if occupied[i][0] > occupied[j][0] {
				occupied[i], occupied[j] = occupied[j], occupied[i]
			}
		}
	}

	// Verificar espacio al inicio
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

	// Buscar entre espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSpace := availableEnd - availableStart

		if availableSpace >= Models.EBR_SIZE+size && availableSpace < bestSize {
			bestStart = availableStart
			bestSize = availableSpace
		}
	}

	// Verificar espacio al final
	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := extEnd - lastEnd
		if finalSpace >= Models.EBR_SIZE+size && finalSpace < bestSize {
			bestStart = lastEnd
		}
	}

	return bestStart
}

// findWorstFitInExtended implementa Worst Fit dentro de la partición extendida
func (e *EBRManager) findWorstFitInExtended(occupied [][2]int64, size int64, extStart int64, extEnd int64) int64 {
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

	// Verificar espacio al inicio
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

	// Buscar entre espacios ocupados
	for i := 0; i < len(occupied)-1; i++ {
		availableStart := occupied[i][1]
		availableEnd := occupied[i+1][0]
		availableSpace := availableEnd - availableStart

		if availableSpace >= Models.EBR_SIZE+size && availableSpace > worstSize {
			worstStart = availableStart
			worstSize = availableSpace
		}
	}

	// Verificar espacio al final
	if len(occupied) > 0 {
		lastEnd := occupied[len(occupied)-1][1]
		finalSpace := extEnd - lastEnd
		if finalSpace >= Models.EBR_SIZE+size && finalSpace > worstSize {
			worstStart = lastEnd
		}
	}

	return worstStart
}

// getEBRForInsertion obtiene el EBR en una posición o el más cercano para inserción
func (e *EBRManager) getEBRForInsertion(position int64) (*Models.EBR, int64, error) {
	// Si la posición es al inicio de la partición extendida, usar el primer EBR
	if position == e.extendedPartition.PartStart {
		ebr, err := e.ReadEBR(position)
		return ebr, position, err
	}

	// Buscar el EBR más cercano antes de esta posición
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

	// Leer el EBR en la última posición encontrada
	ebr, err := e.ReadEBR(lastEBRPos)
	return ebr, lastEBRPos, err
}

// insertNewEBRInChain inserta un nuevo EBR en la cadena existente
func (e *EBRManager) insertNewEBRInChain(name string, size int64, fitType byte, position int64) error {
	// Crear nuevo EBR
	newEBR := Models.EBR{
		PartMount: Models.EBR_UNMOUNTED,
		PartFit:   fitType,
		PartStart: position + Models.EBR_SIZE,
		PartS:     size,
		PartNext:  Models.EBR_END, // Se actualizará si es necesario
	}
	newEBR.SetLogicalPartitionName(name)

	// Encontrar donde insertar en la cadena
	// Esto requiere actualizar los punteros de la lista enlazada
	err := e.updateEBRChainForInsertion(&newEBR, position)
	if err != nil {
		return fmt.Errorf("error actualizando cadena EBRs")
	}

	// Escribir el nuevo EBR
	err = e.WriteEBR(&newEBR, position)
	if err != nil {
		return fmt.Errorf("error escribiendo EBR")
	}

	return nil
}

// updateEBRChainForInsertion actualiza la cadena de EBRs para insertar uno nuevo
func (e *EBRManager) updateEBRChainForInsertion(newEBR *Models.EBR, newPosition int64) error {
	// Recorrer la cadena para encontrar donde insertar
	currentPos := e.extendedPartition.PartStart

	for currentPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR")
		}

		// Si el siguiente EBR está después de nuestra nueva posición
		if ebr.HasNext() && ebr.GetNextEBRPosition() > newPosition {
			// Insertar el nuevo EBR entre el actual y el siguiente
			nextEBRPos := ebr.GetNextEBRPosition()
			ebr.SetNextEBRPosition(newPosition)
			newEBR.SetNextEBRPosition(nextEBRPos)

			// Escribir el EBR actualizado
			err = e.WriteEBR(ebr, currentPos)
			if err != nil {
				return fmt.Errorf("error actualizando EBR")
			}

			return nil
		}

		// Si este es el último EBR y nuestra posición es después
		if !ebr.HasNext() {
			ebr.SetNextEBRPosition(newPosition)
			newEBR.MarkAsLastEBR()

			// Escribir el EBR actualizado
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

	for _, partition := range logicalPartitions {
		if partition.Name == name {
			return true, nil
		}
	}

	return false, nil
}

// GetLogicalPartitions retorna lista de todas las particiones lógicas
func (e *EBRManager) GetLogicalPartitions() ([]Models.LogicalPartitionInfo, error) {
	logicalPartitions := make([]Models.LogicalPartitionInfo, 0)

	// Recorrer la lista enlazada de EBRs
	currentPos := e.extendedPartition.PartStart

	for currentPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return nil, fmt.Errorf("error leyendo EBR")
		}

		// Si el EBR no está vacío, agregarlo a la lista
		if !ebr.IsEmptyEBR() {
			partitionInfo := ebr.ToLogicalPartitionInfo(currentPos)
			logicalPartitions = append(logicalPartitions, partitionInfo)
		}

		// Avanzar al siguiente EBR
		if ebr.HasNext() {
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return logicalPartitions, nil
}

// RemoveLogicalPartition elimina una partición lógica de la cadena de EBRs
func (e *EBRManager) RemoveLogicalPartition(partitionName string) error {
	// Buscar la partición lógica en la cadena
	currentPos := e.extendedPartition.PartStart
	var previousPos int64 = -1

	for currentPos != Models.EBR_END {
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR")
		}

		// Si encontramos la partición a eliminar
		if !ebr.IsEmptyEBR() && ebr.GetLogicalPartitionName() == partitionName {
			return e.removeEBRFromChain(currentPos, previousPos, ebr)
		}

		// Avanzar al siguiente EBR
		if ebr.HasNext() {
			previousPos = currentPos
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return fmt.Errorf("partición lógica '%s' no encontrada", partitionName)
}

// removeEBRFromChain elimina un EBR específico de la cadena
func (e *EBRManager) removeEBRFromChain(ebrPos int64, previousEBRPos int64, ebrToRemove *Models.EBR) error {
	// Si es el primer EBR de la cadena
	if previousEBRPos == -1 {
		// Limpiar el EBR pero mantenerlo como placeholder
		ebrToRemove.ClearEBR()
		return e.WriteEBR(ebrToRemove, ebrPos)
	}

	// Si no es el primer EBR, actualizar el puntero del anterior
	previousEBR, err := e.ReadEBR(previousEBRPos)
	if err != nil {
		return fmt.Errorf("error leyendo EBR anterior")
	}

	// El EBR anterior debe apuntar al siguiente del EBR a eliminar
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

// ValidateEBRChain valida la consistencia de toda la cadena de EBRs
func (e *EBRManager) ValidateEBRChain() error {
	currentPos := e.extendedPartition.PartStart
	visitedPositions := make(map[int64]bool)

	for currentPos != Models.EBR_END {
		// Verificar que no hemos visitado esta posición antes (evitar loops)
		if visitedPositions[currentPos] {
			return fmt.Errorf("loop detectado en la cadena de EBRs en posición %d", currentPos)
		}
		visitedPositions[currentPos] = true

		// Verificar que la posición está dentro de la partición extendida
		if currentPos < e.extendedPartition.PartStart ||
			currentPos >= e.extendedPartition.GetPartitionEnd() {
			return fmt.Errorf("EBR en posición %d está fuera de la partición extendida", currentPos)
		}

		// Leer y validar el EBR
		ebr, err := e.ReadEBR(currentPos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR")
		}

		// Validar el EBR individual
		if !ebr.IsEmptyEBR() {
			err = ebr.ValidateEBR()
			if err != nil {
				return fmt.Errorf("EBR inválido en posición %d: %v", currentPos, err)
			}

			// Verificar que la partición lógica está dentro de la extendida
			if ebr.GetPartitionEnd() > e.extendedPartition.GetPartitionEnd() {
				return fmt.Errorf("partición lógica en EBR %d excede la partición extendida", currentPos)
			}
		}

		// Avanzar al siguiente EBR
		if ebr.HasNext() {
			currentPos = ebr.GetNextEBRPosition()
		} else {
			break
		}
	}

	return nil
}
