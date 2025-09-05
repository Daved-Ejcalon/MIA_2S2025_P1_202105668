package Models

import (
	"unsafe"
)

// Partition representa una entrada de particion en la tabla MBR
type Partition struct {
	PartStatus      byte     // Estado de montaje (0=inactiva, 1=activa)
	PartType        byte     // Tipo: P=primaria, E=extendida, L=logica
	PartFit         byte     // Algoritmo: F=First, B=Best, W=Worst
	PartStart       int64    // Posicion de inicio en bytes
	PartSize        int64    // Tamano de la particion en bytes
	PartName        [16]byte // Nombre de la particion (max 15 caracteres)
	PartCorrelative int64    // Numero correlativo para montaje
	PartID          [4]byte  // ID de montaje asignado
}

// MBR contiene metadatos del disco y tabla de particiones
type MBR struct {
	MbrSize         int64        // Tamano total del disco en bytes
	MbrCreationDate int64        // Timestamp de creacion del disco
	MbrSignature    int64        // Numero aleatorio de identificacion
	DiskFit         byte         // Algoritmo de ajuste por defecto
	Partitions      [4]Partition // Tabla de particiones (max 4 entradas)
}

const (
	PARTITION_PRIMARY  = 'P'
	PARTITION_EXTENDED = 'E'
	PARTITION_LOGICAL  = 'L'
)

const (
	FIT_BEST  = 'B'
	FIT_FIRST = 'F'
	FIT_WORST = 'W'
)

const (
	PARTITION_ACTIVE   = 1
	PARTITION_INACTIVE = 0
)

const (
	MBR_SIZE = 1024
)

func GetMBRSize() int {
	return int(unsafe.Sizeof(MBR{}))
}

func GetPartitionSize() int {
	return int(unsafe.Sizeof(Partition{}))
}

func IsValidPartitionType(partType byte) bool {
	return partType == PARTITION_PRIMARY || partType == PARTITION_EXTENDED || partType == PARTITION_LOGICAL
}

func IsValidFitType(fitType byte) bool {
	return fitType == FIT_BEST || fitType == FIT_FIRST || fitType == FIT_WORST
}

func (p *Partition) IsEmptyPartition() bool {
	return p.PartStart == 0 && p.PartSize == 0
}

func (p *Partition) GetName() string {
	return p.GetPartitionName()
}

func (p *Partition) GetPartitionName() string {
	// Extraer nombre eliminando bytes nulos
	name := make([]byte, 0, 16)
	for _, b := range p.PartName {
		if b == 0 {
			break
		}
		name = append(name, b)
	}
	return string(name)
}

func (p *Partition) SetPartitionName(name string) {
	// Limpiar nombre anterior
	for i := range p.PartName {
		p.PartName[i] = 0
	}

	// Asignar nombre nuevo con limite de 15 caracteres
	nameBytes := []byte(name)
	maxLen := 15
	if len(nameBytes) < maxLen {
		maxLen = len(nameBytes)
	}

	for i := 0; i < maxLen; i++ {
		p.PartName[i] = nameBytes[i]
	}
}

func (p *Partition) GetPartitionID() string {
	id := make([]byte, 0, 4)
	for _, b := range p.PartID {
		if b == 0 {
			break
		}
		id = append(id, b)
	}
	return string(id)
}

func (p *Partition) SetPartitionID(id string) {
	for i := range p.PartID {
		p.PartID[i] = 0
	}

	idBytes := []byte(id)
	maxLen := 4
	if len(idBytes) < maxLen {
		maxLen = len(idBytes)
	}

	for i := 0; i < maxLen; i++ {
		p.PartID[i] = idBytes[i]
	}
}

func (p *Partition) IsPrimary() bool {
	return p.PartType == PARTITION_PRIMARY
}

func (p *Partition) IsExtended() bool {
	return p.PartType == PARTITION_EXTENDED
}

func (p *Partition) IsMounted() bool {
	return p.PartStatus == PARTITION_ACTIVE
}

func (p *Partition) Mount(partitionNumber int64, mountID string) {
	// Marcar particion como activa y asignar ID de montaje
	p.PartStatus = PARTITION_ACTIVE
	p.PartCorrelative = partitionNumber
	copy(p.PartID[:], mountID)
}

func (p *Partition) Unmount() {
	// Desmontar particion y limpiar datos de montaje
	p.PartStatus = PARTITION_INACTIVE
	p.PartCorrelative = -1
	for i := range p.PartID {
		p.PartID[i] = 0
	}
}

func (p *Partition) GetPartitionEnd() int64 {
	return p.PartStart + p.PartSize
}
