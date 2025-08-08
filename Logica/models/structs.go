package models

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
	MbrCreationDate int64 // Unix timestamp
	MbrSignature    int64
	DiskFit         byte
	Partitions      [4]Partition
}
