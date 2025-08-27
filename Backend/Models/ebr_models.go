// Package Models define las estructuras de datos para el sistema EBR
package Models

import (
	"fmt"
	"unsafe"
)

// EBR (Extended Boot Record) estructura para particiones lógicas
// Implementa una lista enlazada donde cada EBR apunta al siguiente
// Reside dentro del espacio de la partición extendida
type EBR struct {
	PartMount byte     // Estado de montaje de la partición lógica
	PartFit   byte     // B (Best) / F (First) / W (Worst)
	PartStart int64    // Byte donde inicia la partición lógica (relativo al disco)
	PartS     int64    // Tamaño total de la partición en bytes
	PartNext  int64    // Byte donde está el próximo EBR (-1 si no hay siguiente)
	PartName  [16]byte // Nombre de la partición lógica (máximo 16 caracteres)
}

// Constantes para estado de montaje del EBR
const (
	EBR_MOUNTED   = 1 // Partición lógica montada
	EBR_UNMOUNTED = 0 // Partición lógica no montada
)

// Constantes del sistema EBR
const (
	EBR_SIZE = 1024 // Tamaño del EBR en bytes (1 KB, similar al MBR)
	EBR_END  = -1   // Valor que indica fin de lista enlazada
)

// GetEBRSize retorna el tamaño en bytes de la estructura EBR
func GetEBRSize() int {
	return int(unsafe.Sizeof(EBR{}))
}

// IsEmptyEBR verifica si un EBR está vacío/sin usar
func (e *EBR) IsEmptyEBR() bool {
	return e.PartStart == 0 && e.PartS == 0
}

// GetLogicalPartitionName retorna el nombre de la partición lógica como string
func (e *EBR) GetLogicalPartitionName() string {
	// Convertir array de bytes a string eliminando null bytes
	name := make([]byte, 0, 16)
	for _, b := range e.PartName {
		if b == 0 {
			break
		}
		name = append(name, b)
	}
	return string(name)
}

// SetLogicalPartitionName establece el nombre de la partición lógica
func (e *EBR) SetLogicalPartitionName(name string) {
	// Limpiar el array
	for i := range e.PartName {
		e.PartName[i] = 0
	}

	// Copiar el nombre limitado a 15 caracteres (dejando espacio para null terminator)
	nameBytes := []byte(name)
	maxLen := 15
	if len(nameBytes) < maxLen {
		maxLen = len(nameBytes)
	}

	for i := 0; i < maxLen; i++ {
		e.PartName[i] = nameBytes[i]
	}
}

// IsMounted verifica si la partición lógica está montada
func (e *EBR) IsMounted() bool {
	return e.PartMount == EBR_MOUNTED
}

// Mount marca la partición lógica como montada
func (e *EBR) Mount() {
	e.PartMount = EBR_MOUNTED
}

// Unmount marca la partición lógica como desmontada
func (e *EBR) Unmount() {
	e.PartMount = EBR_UNMOUNTED
}

// HasNext verifica si hay un siguiente EBR en la lista enlazada
func (e *EBR) HasNext() bool {
	return e.PartNext != EBR_END
}

// GetPartitionEnd retorna la posición donde termina la partición lógica
func (e *EBR) GetPartitionEnd() int64 {
	return e.PartStart + e.PartS
}

// GetNextEBRPosition retorna la posición del siguiente EBR
// Retorna -1 si no hay siguiente EBR
func (e *EBR) GetNextEBRPosition() int64 {
	return e.PartNext
}

// SetNextEBRPosition establece la posición del siguiente EBR
func (e *EBR) SetNextEBRPosition(position int64) {
	e.PartNext = position
}

// MarkAsLastEBR marca este EBR como el último de la lista enlazada
func (e *EBR) MarkAsLastEBR() {
	e.PartNext = EBR_END
}

// IsLastEBR verifica si este es el último EBR en la lista enlazada
func (e *EBR) IsLastEBR() bool {
	return e.PartNext == EBR_END
}

// IsValidFitType verifica si el tipo de ajuste del EBR es válido
func (e *EBR) IsValidFitType() bool {
	return IsValidFitType(e.PartFit)
}

// GetEBROffset retorna el offset donde debería escribirse este EBR
// Esto es útil para calcular posiciones en la lista enlazada
func (e *EBR) GetEBROffset() int64 {
	// El EBR se escribe justo antes de los datos de la partición lógica
	return e.PartStart - EBR_SIZE
}

// ValidateEBR verifica que los campos del EBR sean consistentes
func (e *EBR) ValidateEBR() error {
	// Validar que el tamaño sea positivo si no es EBR vacío
	if !e.IsEmptyEBR() && e.PartS <= 0 {
		return fmt.Errorf("el tamaño de la partición lógica debe ser mayor a 0")
	}

	// Validar tipo de ajuste
	if !e.IsEmptyEBR() && !e.IsValidFitType() {
		return fmt.Errorf("tipo de ajuste inválido: %c", e.PartFit)
	}

	// Validar que PartStart sea válido
	if !e.IsEmptyEBR() && e.PartStart <= 0 {
		return fmt.Errorf("posición de inicio inválida: %d", e.PartStart)
	}

	return nil
}

// ClearEBR limpia todos los campos del EBR (para eliminación)
func (e *EBR) ClearEBR() {
	e.PartMount = EBR_UNMOUNTED
	e.PartFit = 0
	e.PartStart = 0
	e.PartS = 0
	e.PartNext = EBR_END

	// Limpiar nombre
	for i := range e.PartName {
		e.PartName[i] = 0
	}
}

// LogicalPartitionInfo estructura auxiliar para información de partición lógica
type LogicalPartitionInfo struct {
	Name        string // Nombre de la partición lógica
	Start       int64  // Posición de inicio
	Size        int64  // Tamaño en bytes
	IsMounted   bool   // Estado de montaje
	EBRPosition int64  // Posición del EBR que la describe
}

// ToLogicalPartitionInfo convierte un EBR a información de partición lógica
func (e *EBR) ToLogicalPartitionInfo(ebrPosition int64) LogicalPartitionInfo {
	return LogicalPartitionInfo{
		Name:        e.GetLogicalPartitionName(),
		Start:       e.PartStart,
		Size:        e.PartS,
		IsMounted:   e.IsMounted(),
		EBRPosition: ebrPosition,
	}
}
