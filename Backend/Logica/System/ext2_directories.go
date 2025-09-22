package System

import (
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"strings"
)

// EXT2DirectoryManager maneja operaciones de directorios en sistema EXT2
type EXT2DirectoryManager struct {
	fileManager *EXT2FileManager
	manager     *EXT2Manager
}

func NewEXT2DirectoryManager(manager *EXT2Manager) *EXT2DirectoryManager {
	if manager == nil {
		return nil
	}

	fileManager := NewEXT2FileManager(manager)
	if fileManager == nil {
		return nil
	}

	return &EXT2DirectoryManager{
		fileManager: fileManager,
		manager:     manager,
	}
}

// CreateDirectory crea un nuevo directorio con permisos y propietario especificados
func (d *EXT2DirectoryManager) CreateDirectory(dirPath string, uid int32, gid int32, permissions int32) error {
	// Separar ruta padre y nombre del directorio
	parentPath, dirName := d.fileManager.splitPath(dirPath)

	// Verificar que el directorio padre existe
	parentInodeNum, err := d.fileManager.findFileInode(parentPath)
	if err != nil {
		return errors.New("directorio padre no existe")
	}

	parentInodo, err := d.fileManager.readInode(parentInodeNum)
	if err != nil {
		return err
	}

	if parentInodo.I_type != Models.INODO_DIRECTORIO {
		return errors.New("el padre no es un directorio")
	}

	// Los permisos se verifican en la capa de comando (mkdir.go) para evitar ciclos de importación

	// Verificar que el directorio no existe ya
	_, err = d.fileManager.findFileInode(dirPath)
	if err == nil {
		return errors.New("el directorio ya existe")
	}

	return d.createNewDirectory(parentInodeNum, dirName, uid, gid, permissions)
}

// RemoveDirectory elimina un directorio vacío del sistema de archivos
func (d *EXT2DirectoryManager) RemoveDirectory(dirPath string) error {
	// Buscar el inodo del directorio a eliminar
	dirInodeNum, err := d.fileManager.findFileInode(dirPath)
	if err != nil {
		return errors.New("directorio no encontrado")
	}

	// Proteger el directorio raíz
	if dirInodeNum == Models.ROOT_INODE {
		return errors.New("no se puede eliminar el directorio ra�z")
	}

	dirInodo, err := d.fileManager.readInode(dirInodeNum)
	if err != nil {
		return err
	}

	if dirInodo.I_type != Models.INODO_DIRECTORIO {
		return errors.New("no es un directorio")
	}

	isEmpty, err := d.isDirectoryEmpty(dirInodo)
	if err != nil {
		return err
	}

	if !isEmpty {
		return errors.New("el directorio no est� vac�o")
	}

	err = d.freeDirectoryBlocks(dirInodo)
	if err != nil {
		return err
	}

	err = d.fileManager.updateInodeBitmap(dirInodeNum, false)
	if err != nil {
		return err
	}

	return d.removeEntryFromParent(dirPath, dirInodeNum)
}

// ListDirectory retorna lista de entradas del directorio con metadatos
func (d *EXT2DirectoryManager) ListDirectory(dirPath string) ([]DirectoryEntry, error) {
	// Buscar inodo del directorio
	dirInodeNum, err := d.fileManager.findFileInode(dirPath)
	if err != nil {
		return nil, errors.New("directorio no encontrado")
	}

	dirInodo, err := d.fileManager.readInode(dirInodeNum)
	if err != nil {
		return nil, err
	}

	if dirInodo.I_type != Models.INODO_DIRECTORIO {
		return nil, errors.New("no es un directorio")
	}

	var entries []DirectoryEntry

	// Recorrer los 12 bloques directos del directorio
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := d.fileManager.readDirectoryBlock(dirInodo.I_block[i])
		if err != nil {
			return nil, err
		}

		// Procesar cada entrada válida del bloque
		for _, entry := range dirBlock.B_content {
			if entry.B_inodo != Models.FREE_INODE {
				entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")
				if entryName != "" {
					entryInodo, err := d.fileManager.readInode(entry.B_inodo)
					if err != nil {
						continue
					}

					dirEntry := DirectoryEntry{
						Name:        entryName,
						InodeNumber: entry.B_inodo,
						Type:        entryInodo.I_type,
						Size:        entryInodo.I_s,
						Permissions: Models.GetPermissions(entryInodo.I_perm),
						UID:         entryInodo.I_uid,
						GID:         entryInodo.I_gid,
						ATime:       entryInodo.I_atime,
						CTime:       entryInodo.I_ctime,
						MTime:       entryInodo.I_mtime,
					}

					entries = append(entries, dirEntry)
				}
			}
		}
	}

	return entries, nil
}

// createNewDirectory crea directorio asignando inodo y bloque, inicializando con . y ..
func (d *EXT2DirectoryManager) createNewDirectory(parentInodeNum int32, dirName string, uid int32, gid int32, permissions int32) error {
	// Asignar inodo y bloque libre para el nuevo directorio
	newInodeNum, err := d.fileManager.findFreeInode()
	if err != nil {
		return err
	}

	newBlockNum, err := d.fileManager.findFreeBlock()
	if err != nil {
		return err
	}

	newInodo := Models.Inodo{
		I_uid:   uid,
		I_gid:   gid,
		I_s:     Models.BLOQUE_SIZE,
		I_atime: float64(Models.GetCurrentUnixTime()),
		I_ctime: float64(Models.GetCurrentUnixTime()),
		I_mtime: float64(Models.GetCurrentUnixTime()),
		I_type:  Models.INODO_DIRECTORIO,
		I_perm:  Models.SetPermissions(permissions),
	}

	for i := range newInodo.I_block {
		newInodo.I_block[i] = Models.FREE_BLOCK
	}
	newInodo.I_block[0] = newBlockNum

	dirBlock := Models.BloqueCarpeta{}

	for i := range dirBlock.B_content {
		dirBlock.B_content[i].B_inodo = int32(Models.FREE_INODE)
		for j := range dirBlock.B_content[i].B_name {
			dirBlock.B_content[i].B_name[j] = 0
		}
	}

	// Crear entradas estándar: . (directorio actual) y .. (directorio padre)
	dirBlock.B_content[0].B_inodo = int32(newInodeNum)
	copy(dirBlock.B_content[0].B_name[:], ".")

	dirBlock.B_content[1].B_inodo = int32(parentInodeNum)
	copy(dirBlock.B_content[1].B_name[:], "..")

	err = d.fileManager.writeDirectoryBlock(newBlockNum, &dirBlock)
	if err != nil {
		return err
	}

	err = d.fileManager.writeInode(newInodeNum, &newInodo)
	if err != nil {
		return err
	}

	err = d.fileManager.addEntryToDirectory(parentInodeNum, dirName, newInodeNum)
	if err != nil {
		return err
	}

	err = d.fileManager.markInodeAsUsed(newInodeNum)
	if err != nil {
		return err
	}

	err = d.fileManager.markBlockAsUsed(newBlockNum)
	if err != nil {
		return err
	}

	return nil
}

func (d *EXT2DirectoryManager) isDirectoryEmpty(dirInodo *Models.Inodo) (bool, error) {
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := d.fileManager.readDirectoryBlock(dirInodo.I_block[i])
		if err != nil {
			return false, err
		}

		for j, entry := range dirBlock.B_content {
			if entry.B_inodo != Models.FREE_INODE {
				entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")
				if j > 1 && entryName != "." && entryName != ".." && entryName != "" {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func (d *EXT2DirectoryManager) freeDirectoryBlocks(dirInodo *Models.Inodo) error {
	for i := 0; i < 12; i++ {
		if dirInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		err := d.fileManager.updateBlockBitmap(dirInodo.I_block[i], false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *EXT2DirectoryManager) removeEntryFromParent(dirPath string, inodeNum int32) error {
	parentPath, dirName := d.fileManager.splitPath(dirPath)

	parentInodeNum, err := d.fileManager.findFileInode(parentPath)
	if err != nil {
		return err
	}

	parentInodo, err := d.fileManager.readInode(parentInodeNum)
	if err != nil {
		return err
	}

	for i := 0; i < 12; i++ {
		if parentInodo.I_block[i] == Models.FREE_BLOCK {
			break
		}

		dirBlock, err := d.fileManager.readDirectoryBlock(parentInodo.I_block[i])
		if err != nil {
			return err
		}

		for j := 0; j < len(dirBlock.B_content); j++ {
			if dirBlock.B_content[j].B_inodo == inodeNum {
				entryName := strings.TrimRight(string(dirBlock.B_content[j].B_name[:]), "\x00")
				if entryName == dirName {
					dirBlock.B_content[j].B_inodo = int32(Models.FREE_INODE)
					for k := range dirBlock.B_content[j].B_name {
						dirBlock.B_content[j].B_name[k] = 0
					}

					return d.fileManager.writeDirectoryBlock(parentInodo.I_block[i], dirBlock)
				}
			}
		}
	}

	return errors.New("entrada no encontrada en directorio padre")
}

// ChangeDirectory cambia directorio actual manejando rutas absolutas y relativas
func (d *EXT2DirectoryManager) ChangeDirectory(currentDir, targetDir string) (string, error) {
	// Manejar rutas absolutas directamente
	if strings.HasPrefix(targetDir, "/") {
		_, err := d.fileManager.findFileInode(targetDir)
		if err != nil {
			return "", errors.New("directorio no encontrado")
		}
		return targetDir, nil
	}

	var newPath string
	if currentDir == "/" {
		newPath = "/" + targetDir
	} else {
		if targetDir == ".." {
			parts := strings.Split(strings.Trim(currentDir, "/"), "/")
			if len(parts) > 1 {
				newPath = "/" + strings.Join(parts[:len(parts)-1], "/")
			} else {
				newPath = "/"
			}
		} else if targetDir == "." {
			newPath = currentDir
		} else {
			newPath = strings.TrimSuffix(currentDir, "/") + "/" + targetDir
		}
	}

	inodeNum, err := d.fileManager.findFileInode(newPath)
	if err != nil {
		return "", errors.New("directorio no encontrado")
	}

	inodo, err := d.fileManager.readInode(inodeNum)
	if err != nil {
		return "", err
	}

	if inodo.I_type != Models.INODO_DIRECTORIO {
		return "", errors.New("no es un directorio")
	}

	return newPath, nil
}

func (d *EXT2DirectoryManager) GetDirectoryInfo(dirPath string) (*DirectoryInfo, error) {
	inodeNum, err := d.fileManager.findFileInode(dirPath)
	if err != nil {
		return nil, err
	}

	inodo, err := d.fileManager.readInode(inodeNum)
	if err != nil {
		return nil, err
	}

	if inodo.I_type != Models.INODO_DIRECTORIO {
		return nil, errors.New("no es un directorio")
	}

	entries, err := d.ListDirectory(dirPath)
	if err != nil {
		return nil, err
	}

	info := &DirectoryInfo{
		Path:        dirPath,
		InodeNumber: inodeNum,
		Size:        inodo.I_s,
		Permissions: Models.GetPermissions(inodo.I_perm),
		UID:         inodo.I_uid,
		GID:         inodo.I_gid,
		EntryCount:  int32(len(entries)),
	}

	return info, nil
}

// DirectoryEntry representa una entrada de directorio con metadatos
type DirectoryEntry struct {
	Name        string
	InodeNumber int32
	Type        byte
	Size        int32
	Permissions int32
	UID         int32
	GID         int32
	ATime       float64 // Access time
	CTime       float64 // Creation time
	MTime       float64 // Modification time
}

// DirectoryInfo contiene información completa de un directorio
type DirectoryInfo struct {
	Path        string
	InodeNumber int32
	Size        int32
	Permissions int32
	UID         int32
	GID         int32
	EntryCount  int32
}
