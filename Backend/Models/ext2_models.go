package Models

import (
	"time"
	"unsafe"
)

// SuperBloque contiene metadatos del sistema de archivos EXT2
type SuperBloque struct {
	S_filesystem_type   int32   // Tipo de sistema de archivos (2 = EXT2)
	S_inodes_count      int32   // Total de inodos en el sistema
	S_blocks_count      int32   // Total de bloques en el sistema
	S_free_blocks_count int32   // Bloques libres disponibles
	S_free_inodes_count int32   // Inodos libres disponibles
	S_mtime             float64 // Tiempo de ultima modificacion
	S_umtime            float64 // Tiempo de ultimo montaje
	S_mnt_count         int32   // Numero de montajes
	S_magic             int32   // Numero magico EXT2 (0xEF53)
	S_inode_s           int32   // Tamano de cada inodo
	S_block_s           int32   // Tamano de cada bloque
	S_firts_ino         int32   // Primer inodo libre
	S_first_blo         int32   // Primer bloque libre
	S_bm_inode_start    int32   // Posicion del bitmap de inodos
	S_bm_block_start    int32   // Posicion del bitmap de bloques
	S_inode_start       int32   // Posicion de la tabla de inodos
	S_block_start       int32   // Posicion del area de bloques
}

// Inodo representa un archivo o directorio con metadatos y punteros a bloques
type Inodo struct {
	I_uid   int32     // ID del usuario propietario
	I_gid   int32     // ID del grupo propietario
	I_s     int32     // Tamano en bytes
	I_atime float64   // Tiempo de ultimo acceso
	I_ctime float64   // Tiempo de creacion
	I_mtime float64   // Tiempo de ultima modificacion
	I_block [15]int32 // Punteros a bloques (12 directos + 3 indirectos)
	I_type  byte      // Tipo: 0=directorio, 1=archivo
	I_perm  [3]byte   // Permisos en formato octal (owner, group, others)
}

type BloqueCarpeta struct {
	B_content [4]B_content
}

type B_content struct {
	B_name  [12]byte
	B_inodo int32
}

type BloqueArchivos struct {
	B_content [64]byte
}

// GetContent returns a slice of the content bytes
func (ba *BloqueArchivos) GetContent() []byte {
	return ba.B_content[:]
}

// SetContent copies the provided content into the content array
func (ba *BloqueArchivos) SetContent(content []byte) {
	copy(ba.B_content[:], content)
}

type BloqueContenido struct {
	B_content [64]byte
}

type BloqueApuntadores struct {
	B_pointers [16]int32
}

const (
	SUPERBLOQUE_SIZE = 1024
	INODO_SIZE       = 128
	BLOQUE_SIZE      = 64
	BITMAP_SIZE      = 1024

	INODO_ARCHIVO    = 1
	INODO_DIRECTORIO = 0

	EXT2_MAGIC = 0xEF53

	ROOT_INODE = 0

	FREE_BLOCK = -1
	FREE_INODE = -1
)

func GetSuperBloqueSize() int {
	return int(unsafe.Sizeof(SuperBloque{}))
}

func GetInodoSize() int {
	return int(unsafe.Sizeof(Inodo{}))
}

func GetBloqueSize() int {
	return BLOQUE_SIZE
}

// NewSuperBloque crea un SuperBloque inicializado con layout EXT2
func NewSuperBloque(inodesCount, blocksCount int32) SuperBloque {
	currentTime := float64(time.Now().Unix())

	// Calcular posiciones de estructuras en el disco
	inodeBitmapSize := inodesCount
	blockBitmapSize := blocksCount

	return SuperBloque{
		S_filesystem_type:   2,
		S_inodes_count:      inodesCount,
		S_blocks_count:      blocksCount,
		S_free_blocks_count: blocksCount - 2, // -2 por root y users.txt
		S_free_inodes_count: inodesCount - 2, // -2 por root y users.txt
		S_mtime:             currentTime,
		S_umtime:            0.0,
		S_mnt_count:         1,
		S_magic:             EXT2_MAGIC,
		S_inode_s:           INODO_SIZE,
		S_block_s:           BLOQUE_SIZE,
		S_firts_ino:         1,
		S_first_blo:         1,
		S_bm_inode_start:    SUPERBLOQUE_SIZE,
		S_bm_block_start:    SUPERBLOQUE_SIZE + inodeBitmapSize,
		S_inode_start:       SUPERBLOQUE_SIZE + inodeBitmapSize + blockBitmapSize,
		S_block_start:       SUPERBLOQUE_SIZE + inodeBitmapSize + blockBitmapSize + (inodesCount * INODO_SIZE),
	}
}

// NewRootInodo crea el inodo del directorio raiz con permisos 755
func NewRootInodo() Inodo {
	currentTime := float64(time.Now().Unix())

	// Crear inodo del directorio raiz
	inodo := Inodo{
		I_uid:   1,
		I_gid:   1,
		I_s:     BLOQUE_SIZE,
		I_atime: currentTime,
		I_ctime: currentTime,
		I_mtime: currentTime,
		I_type:  INODO_DIRECTORIO,
		I_perm:  SetPermissions(755),
	}

	for i := range inodo.I_block {
		inodo.I_block[i] = FREE_BLOCK
	}

	inodo.I_block[0] = 0

	return inodo
}

func NewRootDirectory() BloqueCarpeta {
	rootDir := BloqueCarpeta{}

	// Inicializar todas las entradas como vacias
	for i := range rootDir.B_content {
		rootDir.B_content[i].B_inodo = int32(FREE_INODE)
		for j := range rootDir.B_content[i].B_name {
			rootDir.B_content[i].B_name[j] = 0
		}
	}

	// Crear entrada "." (directorio actual)
	rootDir.B_content[0].B_inodo = int32(ROOT_INODE)
	copy(rootDir.B_content[0].B_name[:], ".")

	// Crear entrada ".." (directorio padre)
	rootDir.B_content[1].B_inodo = int32(ROOT_INODE)
	copy(rootDir.B_content[1].B_name[:], "..")

	return rootDir
}

func IsValidInodoType(inodoType byte) bool {
	return inodoType == INODO_ARCHIVO || inodoType == INODO_DIRECTORIO
}

func CreateBitmap(size int) []byte {
	return make([]byte, size)
}

func SetBitmapBit(bitmap []byte, position int) {
	if position < 0 || position >= len(bitmap)*8 {
		return
	}

	// Calcular indices de byte y bit para marcar como usado
	byteIndex := position / 8
	bitIndex := position % 8
	bitmap[byteIndex] |= (1 << (7 - bitIndex))
}

func ClearBitmapBit(bitmap []byte, position int) {
	if position < 0 || position >= len(bitmap)*8 {
		return
	}

	// Calcular indices de byte y bit para marcar como libre
	byteIndex := position / 8
	bitIndex := position % 8
	bitmap[byteIndex] &^= (1 << (7 - bitIndex))
}

func IsBitmapBitSet(bitmap []byte, position int) bool {
	if position < 0 || position >= len(bitmap)*8 {
		return false
	}

	byteIndex := position / 8
	bitIndex := position % 8
	return (bitmap[byteIndex] & (1 << (7 - bitIndex))) != 0
}

func FindFreeBitmapBit(bitmap []byte) int {
	// Buscar el primer bit libre en el bitmap
	for byteIndex, b := range bitmap {
		if b != 0xFF { // Si no todos los bits estan ocupados
			for bitIndex := 0; bitIndex < 8; bitIndex++ {
				if (b & (1 << (7 - bitIndex))) == 0 {
					return byteIndex*8 + bitIndex
				}
			}
		}
	}
	return -1 // No hay bits libres
}

func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}

func SetPermissions(octal int32) [3]byte {
	var perms [3]byte
	// Convertir permisos octales (755) a formato [7,5,5]
	perms[0] = byte((octal / 100) % 10) // Owner
	perms[1] = byte((octal / 10) % 10)  // Group
	perms[2] = byte(octal % 10)         // Others
	return perms
}

func GetPermissions(perms [3]byte) int32 {
	// Convertir formato [7,5,5] a permisos octales (755)
	return int32(perms[0])*100 + int32(perms[1])*10 + int32(perms[2])
}
