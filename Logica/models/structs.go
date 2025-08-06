package models

import "time"

type Partition struct {
	PartStatus      byte     // '0' inactiva, '1' activa
	PartType        byte     // 'P' primaria, 'E' extendida
	PartFit         byte     // 'B', 'F', 'W'
	PartStart       int64    // Byte donde empieza
	PartSize        int64    // Tamaño en bytes
	PartName        [16]byte // Nombre (máximo 16 caracteres)
	PartCorrelative int64    // -1 por defecto, luego 1 en adelante al montar
	PartID          [4]byte  // ID de la partición cuando se monta
}

type MBR struct {
	MbrSize         int64
	MbrCreationDate time.Time
	MbrSignature    int64
	DiskFit         byte
	Partitions      [4]Partition
}
