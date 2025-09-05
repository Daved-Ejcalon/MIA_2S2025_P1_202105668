package Models

import (
	"fmt"
	"unsafe"
)

// EBR (Extended Boot Record) maneja particiones logicas en lista enlazada
type EBR struct {
	PartMount byte     // Estado de montaje (0=desmontado, 1=montado)
	PartFit   byte     // Algoritmo de ajuste (B=Best, F=First, W=Worst)
	PartStart int64    // Posicion de inicio de la particion
	PartS     int64    // Tamaño de la particion
	PartNext  int64    // Posicion del siguiente EBR (-1 si es el ultimo)
	PartName  [16]byte // Nombre de la particion (max 15 caracteres)
}

const (
	EBR_MOUNTED   = 1
	EBR_UNMOUNTED = 0
)

const (
	EBR_SIZE = 1024
	EBR_END  = -1
)

func GetEBRSize() int {
	return int(unsafe.Sizeof(EBR{}))
}

// IsEmptyEBR verifica si el EBR esta vacio (sin particion asignada)
func (e *EBR) IsEmptyEBR() bool {
	return e.PartStart == 0 && e.PartS == 0
}

// GetLogicalPartitionName extrae el nombre como string eliminando bytes nulos
func (e *EBR) GetLogicalPartitionName() string {
	name := make([]byte, 0, 16)
	for _, b := range e.PartName {
		if b == 0 {
			break
		}
		name = append(name, b)
	}
	return string(name)
}

// SetLogicalPartitionName asigna nombre limitado a 15 caracteres
func (e *EBR) SetLogicalPartitionName(name string) {
	// Limpiar nombre anterior
	for i := range e.PartName {
		e.PartName[i] = 0
	}

	// Copiar nombre nuevo con limite de 15 caracteres
	nameBytes := []byte(name)
	maxLen := 15
	if len(nameBytes) < maxLen {
		maxLen = len(nameBytes)
	}

	for i := 0; i < maxLen; i++ {
		e.PartName[i] = nameBytes[i]
	}
}

func (e *EBR) IsMounted() bool {
	return e.PartMount == EBR_MOUNTED
}

func (e *EBR) Mount() {
	e.PartMount = EBR_MOUNTED
}

func (e *EBR) Unmount() {
	e.PartMount = EBR_UNMOUNTED
}

// HasNext verifica si existe un EBR siguiente en la cadena
func (e *EBR) HasNext() bool {
	return e.PartNext != EBR_END
}

// GetPartitionEnd calcula la posicion final de la particion
func (e *EBR) GetPartitionEnd() int64 {
	return e.PartStart + e.PartS
}

func (e *EBR) GetNextEBRPosition() int64 {
	return e.PartNext
}

func (e *EBR) SetNextEBRPosition(position int64) {
	e.PartNext = position
}

func (e *EBR) MarkAsLastEBR() {
	e.PartNext = EBR_END
}

func (e *EBR) IsLastEBR() bool {
	return e.PartNext == EBR_END
}

func (e *EBR) IsValidFitType() bool {
	return IsValidFitType(e.PartFit)
}

// GetEBROffset calcula posicion del EBR (antes de la particion)
func (e *EBR) GetEBROffset() int64 {
	return e.PartStart - EBR_SIZE
}

// ValidateEBR verifica integridad de los datos del EBR
func (e *EBR) ValidateEBR() error {
	// Solo validar si el EBR no esta vacio
	if !e.IsEmptyEBR() && e.PartS <= 0 {
		return fmt.Errorf("el tamano de la particion logica debe ser mayor a 0")
	}

	if !e.IsEmptyEBR() && !e.IsValidFitType() {
		return fmt.Errorf("algoritmo de ajuste invalido: %c", e.PartFit)
	}

	if !e.IsEmptyEBR() && e.PartStart <= 0 {
		return fmt.Errorf("posicion de inicio invalida: %d", e.PartStart)
	}

	return nil
}

// ClearEBR limpia completamente el EBR dejandolo vacio
func (e *EBR) ClearEBR() {
	e.PartMount = EBR_UNMOUNTED
	e.PartFit = 0
	e.PartStart = 0
	e.PartS = 0
	e.PartNext = EBR_END

	// Limpiar nombre de la particion
	for i := range e.PartName {
		e.PartName[i] = 0
	}
}

// LogicalPartitionInfo contiene informacion resumida de particion logica
type LogicalPartitionInfo struct {
	Name        string // Nombre de la particion
	Start       int64  // Posicion de inicio
	Size        int64  // Tamaño en bytes
	IsMounted   bool   // Estado de montaje
	EBRPosition int64  // Posicion del EBR en disco
}

// ToLogicalPartitionInfo convierte EBR a estructura de informacion
func (e *EBR) ToLogicalPartitionInfo(ebrPosition int64) LogicalPartitionInfo {
	return LogicalPartitionInfo{
		Name:        e.GetLogicalPartitionName(),
		Start:       e.PartStart,
		Size:        e.PartS,
		IsMounted:   e.IsMounted(),
		EBRPosition: ebrPosition,
	}
}
