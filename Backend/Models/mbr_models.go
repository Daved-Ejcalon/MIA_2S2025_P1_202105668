package Models

import (
	"unsafe"
)

type Partition struct {
	PartStatus      byte
	PartType        byte
	PartFit         byte
	PartStart       int64
	PartSize        int64
	PartName        [16]byte
	PartCorrelative int64
	PartID          [4]byte
}

type MBR struct {
	MbrSize         int64
	MbrCreationDate int64
	MbrSignature    int64
	DiskFit         byte
	Partitions      [4]Partition
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
	for i := range p.PartName {
		p.PartName[i] = 0
	}

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
	p.PartStatus = PARTITION_ACTIVE
	p.PartCorrelative = partitionNumber
	copy(p.PartID[:], mountID)
}

func (p *Partition) Unmount() {
	p.PartStatus = PARTITION_INACTIVE
	p.PartCorrelative = -1
	for i := range p.PartID {
		p.PartID[i] = 0
	}
}

func (p *Partition) GetPartitionEnd() int64 {
	return p.PartStart + p.PartSize
}
