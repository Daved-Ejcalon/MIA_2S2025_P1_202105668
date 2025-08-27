package Models

import (
	"time"
	"unsafe"
)

type SuperBloque struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             float64
	S_umtime            float64
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

type Inodo struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime float64
	I_ctime float64
	I_mtime float64
	I_block [15]int32 // 12 directos, 3 indirectos
	I_type  byte      // '0' archivo, '1' directorio
	I_perm  int32
}

type BloqueCarpeta struct {
	B_content [4]FolderReference
}

type FolderReference struct {
	B_inodo int32
	B_name  [12]byte
}

type BloqueArchivos struct {
	B_content [64]byte
}

type BloqueContenido struct {
	B_content [64]byte
}

const (
	SUPERBLOQUE_SIZE = 1024
	INODO_SIZE       = 128
	BLOQUE_SIZE      = 64
	BITMAP_SIZE      = 1024

	INODO_ARCHIVO    = '0'
	INODO_DIRECTORIO = '1'

	EXT2_MAGIC = 0xEF53

	ROOT_INODE = 0

	FREE_BLOCK = -1
	FREE_INODE = -1
)

// GetSuperBloqueSize retorna el tamaño en bytes de la estructura SuperBloque
func GetSuperBloqueSize() int {
	return int(unsafe.Sizeof(SuperBloque{}))
}

// GetInodoSize retorna el tamaño en bytes de la estructura Inodo
func GetInodoSize() int {
	return int(unsafe.Sizeof(Inodo{}))
}

// GetBloqueSize retorna el tamaño en bytes de un bloque
func GetBloqueSize() int {
	return BLOQUE_SIZE
}

// NewSuperBloque crea un nuevo SuperBloque inicializado con tamaños dinámicos
func NewSuperBloque(inodesCount, blocksCount int32) SuperBloque {
	currentTime := float64(time.Now().Unix())
	
	// Calcular tamaños dinámicos de bitmaps
	inodeBitmapSize := int32(inodesCount/8 + 1)
	blockBitmapSize := int32(blocksCount/8 + 1)
	
	return SuperBloque{
		S_filesystem_type:   2, // EXT2
		S_inodes_count:      inodesCount,
		S_blocks_count:      blocksCount,
		S_free_blocks_count: blocksCount - 1, // -1 por el bloque del directorio raíz
		S_free_inodes_count: inodesCount - 1, // -1 por el inodo raíz
		S_mtime:             currentTime,
		S_umtime:            0.0, // No desmontado aún
		S_mnt_count:         1,   // Primer montaje
		S_magic:             EXT2_MAGIC,
		S_inode_size:        INODO_SIZE,
		S_block_size:        BLOQUE_SIZE,
		S_first_ino:         1,   // Inodo 0 reservado para raíz
		S_first_blo:         1,   // Bloque 0 reservado para raíz
		S_bm_inode_start:    SUPERBLOQUE_SIZE,
		S_bm_block_start:    SUPERBLOQUE_SIZE + inodeBitmapSize,
		S_inode_start:       SUPERBLOQUE_SIZE + inodeBitmapSize + blockBitmapSize,
		S_block_start:       SUPERBLOQUE_SIZE + inodeBitmapSize + blockBitmapSize + (inodesCount * INODO_SIZE),
	}
}

// NewRootInodo crea el inodo del directorio raíz
func NewRootInodo() Inodo {
	currentTime := float64(time.Now().Unix())
	
	inodo := Inodo{
		I_uid:   1,        // Usuario root
		I_gid:   1,        // Grupo root  
		I_size:  BLOQUE_SIZE, // Tamaño inicial del directorio
		I_atime: currentTime,
		I_ctime: currentTime,
		I_mtime: currentTime,
		I_type:  INODO_DIRECTORIO,
		I_perm:  0755, // rwxr-xr-x
	}
	
	// Inicializar punteros a bloques como libres
	for i := range inodo.I_block {
		inodo.I_block[i] = FREE_BLOCK
	}
	
	// Asignar primer bloque al directorio raíz
	inodo.I_block[0] = 0 // Bloque 0 para el directorio raíz
	
	return inodo
}

// NewRootDirectory crea el contenido inicial del directorio raíz
func NewRootDirectory() BloqueCarpeta {
	rootDir := BloqueCarpeta{}
	
	// Inicializar todas las referencias como libres
	for i := range rootDir.B_content {
		rootDir.B_content[i].B_inodo = FREE_INODE
		// Limpiar nombres
		for j := range rootDir.B_content[i].B_name {
			rootDir.B_content[i].B_name[j] = 0
		}
	}
	
	// Entrada para directorio actual (.)
	rootDir.B_content[0].B_inodo = ROOT_INODE
	copy(rootDir.B_content[0].B_name[:], ".")
	
	// Entrada para directorio padre (..)
	rootDir.B_content[1].B_inodo = ROOT_INODE // En raíz, el padre es él mismo
	copy(rootDir.B_content[1].B_name[:], "..")
	
	return rootDir
}

// IsValidInodoType verifica si el tipo de inodo es válido
func IsValidInodoType(inodoType byte) bool {
	return inodoType == INODO_ARCHIVO || inodoType == INODO_DIRECTORIO
}

// CreateBitmap crea un bitmap inicializado con todos los bits en 0 (libres)
func CreateBitmap(size int) []byte {
	return make([]byte, size)
}

// SetBitmapBit marca un bit como ocupado (1) en el bitmap
func SetBitmapBit(bitmap []byte, position int) {
	if position < 0 || position >= len(bitmap)*8 {
		return
	}
	
	byteIndex := position / 8
	bitIndex := position % 8
	bitmap[byteIndex] |= (1 << (7 - bitIndex))
}

// ClearBitmapBit marca un bit como libre (0) en el bitmap
func ClearBitmapBit(bitmap []byte, position int) {
	if position < 0 || position >= len(bitmap)*8 {
		return
	}
	
	byteIndex := position / 8
	bitIndex := position % 8
	bitmap[byteIndex] &^= (1 << (7 - bitIndex))
}

// IsBitmapBitSet verifica si un bit está ocupado (1) en el bitmap
func IsBitmapBitSet(bitmap []byte, position int) bool {
	if position < 0 || position >= len(bitmap)*8 {
		return false
	}
	
	byteIndex := position / 8
	bitIndex := position % 8
	return (bitmap[byteIndex] & (1 << (7 - bitIndex))) != 0
}

// FindFreeBitmapBit encuentra el primer bit libre (0) en el bitmap
func FindFreeBitmapBit(bitmap []byte) int {
	for byteIndex, b := range bitmap {
		if b != 0xFF { // No todos los bits están ocupados
			for bitIndex := 0; bitIndex < 8; bitIndex++ {
				if (b & (1 << (7 - bitIndex))) == 0 {
					return byteIndex*8 + bitIndex
				}
			}
		}
	}
	return -1 // No hay bits libres
}