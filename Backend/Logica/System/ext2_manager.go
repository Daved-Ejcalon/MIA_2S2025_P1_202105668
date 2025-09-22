package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

// MountInfo almacena informacion de una particion montada
type MountInfo struct {
	DiskPath      string
	PartitionName string
	MountID       string
	DiskLetter    rune
	PartNumber    int
}

// EXT2Manager maneja operaciones del sistema de archivos EXT2
type EXT2Manager struct {
	mountInfo     *MountInfo
	diskPath      string
	partitionInfo *Models.Partition
	superBloque   *Models.SuperBloque
}

func NewEXT2Manager(mountInfo *MountInfo) *EXT2Manager {
	manager := &EXT2Manager{
		mountInfo: mountInfo,
		diskPath:  mountInfo.DiskPath,
	}

	// Cargar información de la partición y superbloque
	err := manager.LoadPartitionInfo()
	if err != nil {
		return nil
	}

	err = manager.LoadSuperBlock()
	if err != nil {
		return nil
	}

	return manager
}

// FormatPartition inicializa una particion con sistema de archivos EXT2 completo
func (e *EXT2Manager) FormatPartition() error {
	// Cargar metadatos de la particion desde el MBR
	err := e.LoadPartitionInfo()
	if err != nil {
		return err
	}

	// Calcular distribucion del espacio EXT2
	err = e.calculateEXT2Layout()
	if err != nil {
		return err
	}

	// Escribir SuperBloque con metadatos del sistema
	err = e.writeSuperBloque()
	if err != nil {
		return err
	}

	// Inicializar bitmaps de inodos y bloques
	err = e.initializeBitmaps()
	if err != nil {
		return err
	}

	// Crear directorio raiz con entradas . y ..
	err = e.createRootDirectory()
	if err != nil {
		return err
	}

	// Crear archivo users.txt con usuario root
	err = e.createUsersFile()
	if err != nil {
		return err
	}

	return nil
}

// LoadPartitionInfo carga metadatos de la particion desde el MBR
func (e *EXT2Manager) LoadPartitionInfo() error {
	file, err := os.Open(e.diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer MBR desde el inicio del disco
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return err
	}

	// Buscar particion especifica en tabla MBR
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].GetPartitionName() == e.mountInfo.PartitionName {
			e.partitionInfo = &mbr.Partitions[i]
			break
		}
	}

	if e.partitionInfo == nil {
		return errors.New("particion no encontrada")
	}

	return nil
}

func (e *EXT2Manager) LoadSuperBlock() error {
	if e.superBloque != nil {
		return nil
	}

	file, err := os.Open(e.diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(e.partitionInfo.PartStart, 0)
	if err != nil {
		return err
	}

	var sb Models.SuperBloque
	err = binary.Read(file, binary.LittleEndian, &sb)
	if err != nil {
		return err
	}

	e.superBloque = &sb
	return nil
}

// calculateEXT2Layout calcula distribucion optima de estructuras EXT2
func (e *EXT2Manager) calculateEXT2Layout() error {
	partitionSize := e.partitionInfo.PartSize

	// Obtener tamanos de estructuras EXT2
	superBlockSize := int64(Models.GetSuperBloqueSize())
	inodoSize := int64(Models.GetInodoSize())
	blockSize := int64(Models.GetBloqueSize())

	// Formula: n = (tamano_particion - superbloque) / (4 + inodo + 3*bloque)
	numerator := float64(partitionSize - superBlockSize)
	denominator := float64(4 + inodoSize + 3*blockSize)
	n := numerator / denominator

	// Calcular cantidad de inodos (minimo 1)
	inodesCount := int32(n)
	if inodesCount < 1 {
		return errors.New("particion demasiado pequeña")
	}

	// Relacion 3:1 bloques por inodo
	blocksCount := 3 * inodesCount

	e.superBloque = &Models.SuperBloque{}
	*e.superBloque = Models.NewSuperBloque(inodesCount, blocksCount)

	return nil
}

func (e *EXT2Manager) writeSuperBloque() error {
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(e.partitionInfo.PartStart, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, e.superBloque)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// initializeBitmaps crea y escribe bitmaps iniciales de inodos y bloques
func (e *EXT2Manager) initializeBitmaps() error {
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Crear bitmap de inodos y marcar root (0) y users.txt (1) como usados
	inodeBitmapSize := int(e.superBloque.S_inodes_count)
	inodeBitmap := Models.CreateBitmap(inodeBitmapSize)
	Models.SetBitmapBit(inodeBitmap, 0)
	Models.SetBitmapBit(inodeBitmap, 1)

	inodeBitmapPos := e.partitionInfo.PartStart + int64(e.superBloque.S_bm_inode_start)
	_, err = file.Seek(inodeBitmapPos, 0)
	if err != nil {
		return err
	}
	_, err = file.Write(inodeBitmap)
	if err != nil {
		return err
	}

	// Crear bitmap de bloques y marcar directorio root (0) y users.txt (1) como usados
	blockBitmapSize := int(e.superBloque.S_blocks_count)
	blockBitmap := Models.CreateBitmap(blockBitmapSize)
	Models.SetBitmapBit(blockBitmap, 0)
	Models.SetBitmapBit(blockBitmap, 1)

	blockBitmapPos := e.partitionInfo.PartStart + int64(e.superBloque.S_bm_block_start)
	_, err = file.Seek(blockBitmapPos, 0)
	if err != nil {
		return err
	}
	_, err = file.Write(blockBitmap)
	if err != nil {
		return err
	}

	return nil
}

// createRootDirectory crea directorio raiz con inodo 0 y bloque 0
func (e *EXT2Manager) createRootDirectory() error {
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Crear y escribir inodo del directorio raiz
	rootInodo := Models.NewRootInodo()
	inodoPos := e.partitionInfo.PartStart + int64(e.superBloque.S_inode_start) + int64(Models.ROOT_INODE*Models.INODO_SIZE)

	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &rootInodo)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	// Crear y escribir bloque del directorio raiz con entradas . y ..
	rootDir := Models.NewRootDirectory()
	blockPos := e.partitionInfo.PartStart + int64(e.superBloque.S_block_start) + int64(0*Models.BLOQUE_SIZE)

	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	buffer = new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &rootDir)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// createUsersFile crea archivo users.txt con usuario root predeterminado
func (e *EXT2Manager) createUsersFile() error {
	// Contenido inicial usando la funcion de Models
	usersContent := Models.CreateInitialUsersContent()

	// Crear inodo del archivo users.txt
	usersInodo := Models.Inodo{
		I_uid:   1,
		I_gid:   1,
		I_s:     int32(len(usersContent)),
		I_atime: float64(Models.GetCurrentUnixTime()),
		I_ctime: float64(Models.GetCurrentUnixTime()),
		I_mtime: float64(Models.GetCurrentUnixTime()),
		I_type:  Models.INODO_ARCHIVO,
		I_perm:  Models.SetPermissions(644),
	}

	for i := range usersInodo.I_block {
		usersInodo.I_block[i] = Models.FREE_BLOCK
	}
	usersInodo.I_block[0] = 100  // Usar bloque alto para evitar conflictos

	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	inodoPos := e.partitionInfo.PartStart + int64(e.superBloque.S_inode_start) + int64(1*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &usersInodo)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	blockPos := e.partitionInfo.PartStart + int64(e.superBloque.S_block_start) + int64(100*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	contentBlock := Models.BloqueArchivos{}
	contentBlock.SetContent([]byte(usersContent))

	buffer = new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &contentBlock)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return e.addFileToRootDirectory("users.txt", 1)
}

// addFileToRootDirectory agrega una entrada de archivo al directorio raiz
func (e *EXT2Manager) addFileToRootDirectory(filename string, inodoNumber int32) error {
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Leer bloque del directorio raiz
	blockPos := e.partitionInfo.PartStart + int64(e.superBloque.S_block_start) + int64(0*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	var rootDir Models.BloqueCarpeta
	err = binary.Read(file, binary.LittleEndian, &rootDir)
	if err != nil {
		return err
	}

	// Buscar primera entrada libre (indices 0,1 son . y ..)
	for i := 2; i < len(rootDir.B_content); i++ {
		if rootDir.B_content[i].B_inodo == Models.FREE_INODE {
			rootDir.B_content[i].B_inodo = inodoNumber
			copy(rootDir.B_content[i].B_name[:], filename)
			break
		}
	}

	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &rootDir)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
