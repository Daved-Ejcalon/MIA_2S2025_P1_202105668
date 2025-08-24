// Package Models define las estructuras de datos para el sistema MBR
package Models

import (
	"unsafe"
)

// Partition representa una partición en el MBR
// Contiene la información de cada una de las 4 particiones posibles
type Partition struct {
	PartStatus      byte     // Estado de montaje de la partición
	PartType        byte     // P (Primaria) / E (Extendida)
	PartFit         byte     // B (Best) / F (First) / W (Worst)
	PartStart       int64    // Byte donde inicia la partición en el disco
	PartSize        int64    // Tamaño total de la partición en bytes
	PartName        [16]byte // Nombre de la partición (máximo 16 caracteres)
	PartCorrelative int64    // Correlativo de montaje (-1 = no montada, 1+ = montada)
	PartID          [4]byte  // ID generado al montar la partición
}

// MBR (Master Boot Record) estructura principal del primer sector del disco
// Ocupa exactamente 1024 bytes en el offset 0 del archivo de disco
type MBR struct {
	MbrSize         int64        // Tamaño total del disco en bytes
	MbrCreationDate int64        // Fecha y hora de creación del disco (timestamp Unix)
	MbrSignature    int64        // Número random que identifica únicamente el disco
	DiskFit         byte         // Tipo de ajuste de partición: B (Best), F (First), W (Worst)
	Partitions      [4]Partition // Array de 4 particiones (estándar MBR)
}

// Constantes para tipos de partición
const (
	PARTITION_PRIMARY  = 'P' // Partición primaria
	PARTITION_EXTENDED = 'E' // Partición extendida
	PARTITION_LOGICAL  = 'L' // Partición lógica
)

// Constantes para tipos de ajuste (fit)
const (
	FIT_BEST  = 'B' // Best fit - busca el espacio que mejor se ajuste
	FIT_FIRST = 'F' // First fit - usa el primer espacio disponible
	FIT_WORST = 'W' // Worst fit - usa el espacio más grande disponible
)

// Constantes para estado de partición
const (
	PARTITION_ACTIVE   = 1 // Partición activa/montada
	PARTITION_INACTIVE = 0 // Partición inactiva/no montada
)

// Constantes del sistema
const (
	MBR_SIZE = 1024 // Tamaño del MBR en bytes (1 KB)
)

// GetMBRSize retorna el tamaño en bytes de la estructura MBR
func GetMBRSize() int {
	return int(unsafe.Sizeof(MBR{}))
}

// GetPartitionSize retorna el tamaño en bytes de la estructura Partition
func GetPartitionSize() int {
	return int(unsafe.Sizeof(Partition{}))
}

// IsValidPartitionType verifica si el tipo de partición es válido
func IsValidPartitionType(partType byte) bool {
	return partType == PARTITION_PRIMARY || partType == PARTITION_EXTENDED || partType == PARTITION_LOGICAL
}

// IsValidFitType verifica si el tipo de ajuste es válido
func IsValidFitType(fitType byte) bool {
	return fitType == FIT_BEST || fitType == FIT_FIRST || fitType == FIT_WORST
}

// IsEmptyPartition verifica si una partición está vacía/sin usar
func (p *Partition) IsEmptyPartition() bool {
	return p.PartStart == 0 && p.PartSize == 0
}

// GetName retorna el nombre de la partición como string (alias para GetPartitionName)
func (p *Partition) GetName() string {
	return p.GetPartitionName()
}

// GetPartitionName retorna el nombre de la partición como string
func (p *Partition) GetPartitionName() string {
	// Convertir array de bytes a string eliminando null bytes
	name := make([]byte, 0, 16)
	for _, b := range p.PartName {
		if b == 0 {
			break
		}
		name = append(name, b)
	}
	return string(name)
}

// SetPartitionName establece el nombre de la partición
func (p *Partition) SetPartitionName(name string) {
	// Limpiar el array
	for i := range p.PartName {
		p.PartName[i] = 0
	}

	// Copiar el nombre limitado a 15 caracteres (dejando espacio para null terminator)
	nameBytes := []byte(name)
	maxLen := 15
	if len(nameBytes) < maxLen {
		maxLen = len(nameBytes)
	}

	for i := 0; i < maxLen; i++ {
		p.PartName[i] = nameBytes[i]
	}
}

// GetPartitionID retorna el ID de la partición como string
func (p *Partition) GetPartitionID() string {
	// Convertir array de bytes a string eliminando null bytes
	id := make([]byte, 0, 4)
	for _, b := range p.PartID {
		if b == 0 {
			break
		}
		id = append(id, b)
	}
	return string(id)
}

// SetPartitionID establece el ID de la partición
func (p *Partition) SetPartitionID(id string) {
	// Limpiar el array
	for i := range p.PartID {
		p.PartID[i] = 0
	}

	// Copiar el ID limitado a 4 caracteres
	idBytes := []byte(id)
	maxLen := 4
	if len(idBytes) < maxLen {
		maxLen = len(idBytes)
	}

	for i := 0; i < maxLen; i++ {
		p.PartID[i] = idBytes[i]
	}
}

// IsPrimary verifica si la partición es de tipo primaria
func (p *Partition) IsPrimary() bool {
	return p.PartType == PARTITION_PRIMARY
}

// IsExtended verifica si la partición es de tipo extendida
func (p *Partition) IsExtended() bool {
	return p.PartType == PARTITION_EXTENDED
}

// IsMounted verifica si la partición está montada
func (p *Partition) IsMounted() bool {
	return p.PartStatus == PARTITION_ACTIVE
}

// Mount marca la partición como montada
func (p *Partition) Mount(partitionNumber int64, mountID string) {
	p.PartStatus = PARTITION_ACTIVE
	p.PartCorrelative = partitionNumber
	copy(p.PartID[:], mountID)
}

// Unmount marca la partición como desmontada
func (p *Partition) Unmount() {
	p.PartStatus = PARTITION_INACTIVE
	p.PartCorrelative = -1
	// Limpiar ID
	for i := range p.PartID {
		p.PartID[i] = 0
	}
}

// GetPartitionEnd retorna la posición donde termina la partición
func (p *Partition) GetPartitionEnd() int64 {
	return p.PartStart + p.PartSize
}
