package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strings"
)

// EXT2FileManager maneja operaciones de archivos en sistema EXT2
type EXT2FileManager struct {
	manager *EXT2Manager
}

func NewEXT2FileManager(manager *EXT2Manager) *EXT2FileManager {
	return &EXT2FileManager{
		manager: manager,
	}
}

// ReadFileContent lee el contenido completo de un archivo desde el sistema EXT2
func (f *EXT2FileManager) ReadFileContent(filePath string) (string, error) {
	// Buscar el inodo del archivo
	inodoNumber, err := f.findFileInode(filePath)
	if err != nil {
		return "", err
	}

	inodo, err := f.readInode(inodoNumber)
	if err != nil {
		return "", err
	}

	if inodo.I_type != Models.INODO_ARCHIVO {
		return "", errors.New("no es un archivo")
	}

	content, err := f.readInodeContent(inodo)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// WriteFileContent escribe contenido a un archivo, creando o sobrescribiendo
func (f *EXT2FileManager) WriteFileContent(filePath string, content string, uid int32, gid int32, permissions int32) error {
	// Separar ruta padre y nombre del archivo
	parentPath, fileName := f.splitPath(filePath)

	// Verificar directorio padre
	parentInodoNum, err := f.findFileInode(parentPath)
	if err != nil {
		return err
	}

	parentInodo, err := f.readInode(parentInodoNum)
	if err != nil {
		return err
	}

	if parentInodo.I_type != Models.INODO_DIRECTORIO {
		return errors.New("directorio padre inv�lido")
	}

	// Determinar si crear nuevo archivo o sobreescribir existente
	existingInodo, err := f.findFileInode(filePath)
	if err == nil {
		return f.overwriteFileContent(existingInodo, content)
	}

	return f.createNewFile(parentInodoNum, fileName, content, uid, gid, permissions)
}

// findFileInode navega la jerarquia de directorios para encontrar un archivo
func (f *EXT2FileManager) findFileInode(filePath string) (int32, error) {
	// Caso especial: directorio raiz
	if filePath == "/" {
		return Models.ROOT_INODE, nil
	}

	// Dividir ruta en partes y empezar desde la raiz
	pathParts := strings.Split(strings.Trim(filePath, "/"), "/")
	currentInode := Models.ROOT_INODE

	// Recorrer cada componente de la ruta
	for _, part := range pathParts {
		if part == "" {
			continue
		}

		inodo, err := f.readInode(int32(currentInode))
		if err != nil {
			return -1, err
		}

		if inodo.I_type != Models.INODO_DIRECTORIO {
			return -1, errors.New("no es un directorio")
		}

		nextInode, err := f.findInDirectory(inodo, part)
		if err != nil {
			return -1, err
		}

		currentInode = int(nextInode)
	}

	return int32(currentInode), nil
}

// findInDirectory busca un archivo especifico dentro de un directorio
func (f *EXT2FileManager) findInDirectory(dirInodo *Models.Inodo, filename string) (int32, error) {
	// Recorrer los 12 bloques directos del directorio
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := f.readDirectoryBlock(dirInodo.I_block[i])
		if err != nil {
			return -1, err
		}

		for _, entry := range dirBlock.B_content {
			if entry.B_inodo != Models.FREE_INODE {
				entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")
				if entryName == filename {
					return entry.B_inodo, nil
				}
			}
		}
	}

	return -1, errors.New("archivo no encontrado")
}

func (f *EXT2FileManager) readInode(inodeNumber int32) (*Models.Inodo, error) {
	file, err := os.Open(f.manager.diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	inodoPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_inode_start) + int64(inodeNumber*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return nil, err
	}

	var inodo Models.Inodo
	err = binary.Read(file, binary.LittleEndian, &inodo)
	if err != nil {
		return nil, err
	}

	return &inodo, nil
}

func (f *EXT2FileManager) readDirectoryBlock(blockNumber int32) (*Models.BloqueCarpeta, error) {
	file, err := os.Open(f.manager.diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blockPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return nil, err
	}

	var dirBlock Models.BloqueCarpeta
	err = binary.Read(file, binary.LittleEndian, &dirBlock)
	if err != nil {
		return nil, err
	}

	return &dirBlock, nil
}

// readInodeContent lee el contenido completo de un inodo desde sus bloques
func (f *EXT2FileManager) readInodeContent(inodo *Models.Inodo) ([]byte, error) {
	content := make([]byte, 0, inodo.I_s)

	// Leer contenido de los 12 bloques directos
	for i := 0; i < 12 && i < len(inodo.I_block); i++ {
		if inodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		blockContent, err := f.readFileBlock(inodo.I_block[i])
		if err != nil {
			return nil, err
		}

		content = append(content, blockContent...)
	}

	// Asegurar que no exceda el tamano declarado del archivo
	if int32(len(content)) > inodo.I_s {
		content = content[:inodo.I_s]
	}

	return content, nil
}

func (f *EXT2FileManager) readFileBlock(blockNumber int32) ([]byte, error) {
	file, err := os.Open(f.manager.diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blockPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return nil, err
	}

	var fileBlock Models.BloqueArchivos
	err = binary.Read(file, binary.LittleEndian, &fileBlock)
	if err != nil {
		return nil, err
	}

	return fileBlock.GetContent(), nil
}

func (f *EXT2FileManager) overwriteFileContent(inodeNumber int32, content string) error {
	inodo, err := f.readInode(inodeNumber)
	if err != nil {
		return err
	}

	inodo.I_s = int32(len(content))
	inodo.I_mtime = float64(Models.GetCurrentUnixTime())
	inodo.I_atime = float64(Models.GetCurrentUnixTime())

	if inodo.I_block[0] != Models.FREE_BLOCK {
		err = f.writeFileBlock(inodo.I_block[0], content)
		if err != nil {
			return err
		}
	}

	return f.writeInode(inodeNumber, inodo)
}

// createNewFile crea un archivo nuevo con inodo y bloque asignados
func (f *EXT2FileManager) createNewFile(parentInodeNum int32, fileName string, content string, uid int32, gid int32, permissions int32) error {
	// Asignar recursos: inodo y bloque libres
	newInodeNum, err := f.findFreeInode()
	if err != nil {
		return err
	}

	newBlockNum, err := f.findFreeBlock()
	if err != nil {
		return err
	}

	// Crear inodo del archivo con metadatos completos
	newInodo := Models.Inodo{
		I_uid:   uid,
		I_gid:   gid,
		I_s:     int32(len(content)),
		I_atime: float64(Models.GetCurrentUnixTime()),
		I_ctime: float64(Models.GetCurrentUnixTime()),
		I_mtime: float64(Models.GetCurrentUnixTime()),
		I_type:  Models.INODO_ARCHIVO,
		I_perm:  Models.SetPermissions(permissions),
	}

	for i := range newInodo.I_block {
		newInodo.I_block[i] = Models.FREE_BLOCK
	}
	newInodo.I_block[0] = newBlockNum

	err = f.writeFileBlock(newBlockNum, content)
	if err != nil {
		return err
	}

	err = f.writeInode(newInodeNum, &newInodo)
	if err != nil {
		return err
	}

	err = f.addEntryToDirectory(parentInodeNum, fileName, newInodeNum)
	if err != nil {
		return err
	}

	err = f.markInodeAsUsed(newInodeNum)
	if err != nil {
		return err
	}

	err = f.markBlockAsUsed(newBlockNum)
	if err != nil {
		return err
	}

	return nil
}

func (f *EXT2FileManager) writeFileBlock(blockNumber int32, content string) error {
	file, err := os.OpenFile(f.manager.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	blockPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	fileBlock := Models.BloqueArchivos{}
	fileBlock.SetContent([]byte(content))

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &fileBlock)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	return err
}

func (f *EXT2FileManager) writeInode(inodeNumber int32, inodo *Models.Inodo) error {
	file, err := os.OpenFile(f.manager.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	inodoPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_inode_start) + int64(inodeNumber*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, inodo)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	return err
}

// findFreeInode busca el primer inodo libre usando bitmap
func (f *EXT2FileManager) findFreeInode() (int32, error) {
	bitmap, err := f.readInodeBitmap()
	if err != nil {
		return -1, err
	}

	// Buscar primer bit libre en el bitmap de inodos
	freeIndex := Models.FindFreeBitmapBit(bitmap)
	if freeIndex == -1 {
		return -1, errors.New("no hay inodos libres")
	}

	return int32(freeIndex), nil
}

// findFreeBlock busca el primer bloque libre usando bitmap
func (f *EXT2FileManager) findFreeBlock() (int32, error) {
	bitmap, err := f.readBlockBitmap()
	if err != nil {
		return -1, err
	}

	// Buscar primer bit libre en el bitmap de bloques
	freeIndex := Models.FindFreeBitmapBit(bitmap)
	if freeIndex == -1 {
		return -1, errors.New("no hay bloques libres")
	}

	return int32(freeIndex), nil
}

func (f *EXT2FileManager) readInodeBitmap() ([]byte, error) {
	file, err := os.Open(f.manager.diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bitmapPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_bm_inode_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return nil, err
	}

	bitmapSize := int(f.manager.superBloque.S_inodes_count/8 + 1)
	bitmap := make([]byte, bitmapSize)
	_, err = file.Read(bitmap)
	if err != nil {
		return nil, err
	}

	return bitmap, nil
}

func (f *EXT2FileManager) readBlockBitmap() ([]byte, error) {
	file, err := os.Open(f.manager.diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bitmapPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_bm_block_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return nil, err
	}

	bitmapSize := int(f.manager.superBloque.S_blocks_count/8 + 1)
	bitmap := make([]byte, bitmapSize)
	_, err = file.Read(bitmap)
	if err != nil {
		return nil, err
	}

	return bitmap, nil
}

func (f *EXT2FileManager) markInodeAsUsed(inodeNumber int32) error {
	return f.updateInodeBitmap(inodeNumber, true)
}

func (f *EXT2FileManager) markBlockAsUsed(blockNumber int32) error {
	return f.updateBlockBitmap(blockNumber, true)
}

func (f *EXT2FileManager) updateInodeBitmap(inodeNumber int32, used bool) error {
	bitmap, err := f.readInodeBitmap()
	if err != nil {
		return err
	}

	if used {
		Models.SetBitmapBit(bitmap, int(inodeNumber))
	} else {
		Models.ClearBitmapBit(bitmap, int(inodeNumber))
	}

	return f.writeInodeBitmap(bitmap)
}

func (f *EXT2FileManager) updateBlockBitmap(blockNumber int32, used bool) error {
	bitmap, err := f.readBlockBitmap()
	if err != nil {
		return err
	}

	if used {
		Models.SetBitmapBit(bitmap, int(blockNumber))
	} else {
		Models.ClearBitmapBit(bitmap, int(blockNumber))
	}

	return f.writeBlockBitmap(bitmap)
}

func (f *EXT2FileManager) writeInodeBitmap(bitmap []byte) error {
	file, err := os.OpenFile(f.manager.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	bitmapPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_bm_inode_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(bitmap)
	return err
}

func (f *EXT2FileManager) writeBlockBitmap(bitmap []byte) error {
	file, err := os.OpenFile(f.manager.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	bitmapPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_bm_block_start)
	_, err = file.Seek(bitmapPos, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(bitmap)
	return err
}

func (f *EXT2FileManager) addEntryToDirectory(dirInodeNum int32, filename string, fileInodeNum int32) error {
	dirInodo, err := f.readInode(dirInodeNum)
	if err != nil {
		return err
	}

	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := f.readDirectoryBlock(dirInodo.I_block[i])
		if err != nil {
			return err
		}

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == Models.FREE_INODE {
				dirBlock.B_content[j].B_inodo = int32(fileInodeNum)
				copy(dirBlock.B_content[j].B_name[:], filename)
				return f.writeDirectoryBlock(dirInodo.I_block[i], dirBlock)
			}
		}
	}

	return errors.New("no hay espacio en el directorio")
}

func (f *EXT2FileManager) writeDirectoryBlock(blockNumber int32, dirBlock *Models.BloqueCarpeta) error {
	file, err := os.OpenFile(f.manager.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	blockPos := f.manager.partitionInfo.PartStart + int64(f.manager.superBloque.S_block_start) + int64(blockNumber*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, dirBlock)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer.Bytes())
	return err
}

// splitPath separa una ruta en directorio padre y nombre de archivo
func (f *EXT2FileManager) splitPath(filePath string) (string, string) {
	filePath = strings.Trim(filePath, "/")
	parts := strings.Split(filePath, "/")

	// Caso especial: archivo en directorio raiz
	if len(parts) <= 1 {
		return "/", parts[0]
	}

	// Separar nombre del archivo y construir ruta del padre
	fileName := parts[len(parts)-1]
	parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")

	return parentPath, fileName
}
