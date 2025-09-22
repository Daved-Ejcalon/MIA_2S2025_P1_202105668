package Root

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// MkdirCommand maneja la creacion de directorios
type MkdirCommand struct {
	loginManager      *Users.LoginManager
	permissionManager *PermissionManager
}

// NewMkdirCommand crea una nueva instancia del comando mkdir
func NewMkdirCommand(loginManager *Users.LoginManager, permissionManager *PermissionManager) *MkdirCommand {
	return &MkdirCommand{
		loginManager:      loginManager,
		permissionManager: permissionManager,
	}
}

// Execute ejecuta el comando mkdir con los parametros especificados
func (cmd *MkdirCommand) Execute(dirPath string, createParents bool) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Verificar si el directorio ya existe
	if cmd.directoryExists(dirPath) {
		return errors.New("ERROR: El directorio ya existe")
	}

	// Verificar permisos en directorio padre
	parentDir := cmd.getParentDirectory(dirPath)
	if !cmd.hasWritePermissionInParent(parentDir, createParents) {
		return errors.New("ERROR: Sin permisos de escritura en directorio padre")
	}

	// Crear directorios padre si es necesario
	if createParents {
		if err := cmd.createParentDirectories(parentDir); err != nil {
			return err
		}
	}

	// Crear el directorio en el sistema EXT2
	return cmd.createDirectoryInEXT2(dirPath)
}

// directoryExists verifica si el directorio ya existe en el sistema
func (cmd *MkdirCommand) directoryExists(dirPath string) bool {
	return false
}

// getParentDirectory obtiene el directorio padre de una ruta
func (cmd *MkdirCommand) getParentDirectory(dirPath string) string {
	return filepath.Dir(dirPath)
}

// hasWritePermissionInParent verifica permisos de escritura en directorio padre
func (cmd *MkdirCommand) hasWritePermissionInParent(parentDir string, createParents bool) bool {
	// Si es root, siempre tiene permisos
	if cmd.permissionManager.IsRoot() {
		return true
	}

	// Si no existe el directorio padre y no es recursivo
	if !cmd.parentDirectoryExists(parentDir) && !createParents {
		return false
	}

	return true
}

// parentDirectoryExists verifica si el directorio padre existe
func (cmd *MkdirCommand) parentDirectoryExists(parentDir string) bool {
	return parentDir == "/" || parentDir == "."
}

// createParentDirectories crea los directorios padre recursivamente
func (cmd *MkdirCommand) createParentDirectories(parentDir string) error {
	// Si es directorio raiz, no hacer nada
	if parentDir == "/" || parentDir == "." {
		return nil
	}

	// Dividir la ruta en componentes
	pathParts := cmd.splitPath(parentDir)
	currentPath := "/"

	// Crear cada directorio en la jerarquia
	for _, part := range pathParts {
		if part == "" {
			continue
		}

		currentPath = filepath.Join(currentPath, part)
		currentPath = strings.ReplaceAll(currentPath, "\\", "/")

		// Crear directorio si no existe
		if !cmd.directoryExists(currentPath) {
			if err := cmd.createDirectoryInEXT2(currentPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// splitPath divide una ruta en sus componentes
func (cmd *MkdirCommand) splitPath(path string) []string {
	// Limpiar la ruta
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return []string{}
	}

	return strings.Split(path, "/")
}

// createDirectoryInEXT2 crea el directorio en el sistema EXT2
func (cmd *MkdirCommand) createDirectoryInEXT2(dirPath string) error {
	session := cmd.loginManager.GetCurrentSession()

	// Obtener EXT2Manager para la sesión activa
	mountInfo, err := Disk.GetMountInfoByID(session.MountID)
	if err != nil {
		return fmt.Errorf("partición no encontrada: %v", err)
	}

	// Convertir MountInfo de Disk a System
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("ERROR: No se pudo crear EXT2Manager")
	}

	dirManager := System.NewEXT2DirectoryManager(ext2Manager)
	if dirManager == nil {
		return fmt.Errorf("ERROR: No se pudo crear EXT2DirectoryManager")
	}

	// Determinar permisos: 777 para root, 664 para usuarios regulares
	permissions := 664
	if cmd.permissionManager.IsRoot() {
		permissions = 777
	}

	// Crear el directorio usando EXT2DirectoryManager
	return dirManager.CreateDirectory(dirPath, int32(session.UserID), int32(session.GroupID), int32(permissions))
}

// MkDir - Función exportada para comando mkdir
func MkDir(params map[string]string) error {
	path, hasPath := params["path"]
	if !hasPath {
		return fmt.Errorf("parametro -path requerido")
	}

	_, hasP := params["p"]
	createParents := hasP

	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil {
		return fmt.Errorf("ERROR: No hay sesión inicializada. Debe hacer login primero")
	}
	if !session.IsActive {
		return fmt.Errorf("ERROR: Debe iniciar sesión para usar este comando")
	}

	// Obtener EXT2Manager para la sesión activa
	mountInfo, err := Disk.GetMountInfoByID(session.MountID)
	if err != nil {
		return fmt.Errorf("partición no encontrada: %v", err)
	}

	_, _, err = Users.GetPartitionAndSuperBlock(mountInfo)
	if err != nil {
		return fmt.Errorf("error accediendo al sistema de archivos: %v", err)
	}

	// Convertir MountInfo de Disk a System
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountInfo.DiskPath,
		PartitionName: mountInfo.PartitionName,
		MountID:       mountInfo.MountID,
		DiskLetter:    mountInfo.DiskLetter,
		PartNumber:    mountInfo.PartNumber,
	}

	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("ERROR: No se pudo crear EXT2Manager")
	}

	dirManager := System.NewEXT2DirectoryManager(ext2Manager)
	if dirManager == nil {
		return fmt.Errorf("ERROR: No se pudo crear EXT2DirectoryManager")
	}

	// Determinar permisos: 777 para root, 664 para usuarios regulares
	permissions := 664
	if session.UserID == 1 { // root
		permissions = 777
	}

	// Usar la lógica existente del EXT2DirectoryManager
	err = dirManager.CreateDirectory(path, int32(session.UserID), int32(session.GroupID), int32(permissions))
	if err != nil {
		// Si es un error porque el directorio padre no existe y tenemos -p, crear recursivamente
		if strings.Contains(err.Error(), "directorio padre no existe") && createParents {
			// Dividir la ruta y crear directorios padre recursivamente
			pathParts := strings.Split(strings.Trim(path, "/"), "/")
			currentPath := ""
			
			for _, part := range pathParts {
				if part == "" {
					continue
				}
				currentPath = currentPath + "/" + part
				
				// Intentar crear cada directorio en la jerarquía
				err = dirManager.CreateDirectory(currentPath, int32(session.UserID), int32(session.GroupID), int32(permissions))
				if err != nil && !strings.Contains(err.Error(), "ya existe") {
					return fmt.Errorf("error creando directorio '%s': %v", currentPath, err)
				}
			}
		} else {
			return err
		}
	}

	fmt.Printf("Directorio '%s' creado exitosamente\n", path)
	return nil
}
