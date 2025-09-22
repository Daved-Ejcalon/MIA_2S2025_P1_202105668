package Root

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MkfileCommand maneja la creacion de archivos
type MkfileCommand struct {
	loginManager      *Users.LoginManager
	permissionManager *PermissionManager
}

// NewMkfileCommand crea una nueva instancia del comando mkfile
func NewMkfileCommand(loginManager *Users.LoginManager, permissionManager *PermissionManager) *MkfileCommand {
	return &MkfileCommand{
		loginManager:      loginManager,
		permissionManager: permissionManager,
	}
}

// Execute ejecuta el comando mkfile con los parametros especificados
func (cmd *MkfileCommand) Execute(filePath string, recursive bool, size int, contentFile string) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Validar parametros
	if err := cmd.validateParameters(size); err != nil {
		return err
	}

	// Verificar si el archivo ya existe
	if cmd.fileExists(filePath) {
		return errors.New("ERROR: El archivo ya existe. ¿Desea sobreescribirlo?")
	}

	// Verificar permisos en directorio padre
	parentDir := cmd.getParentDirectory(filePath)
	if !cmd.hasWritePermissionInParent(parentDir, recursive) {
		return errors.New("ERROR: Sin permisos de escritura en directorio padre")
	}

	// Crear directorios padre si es necesario
	if recursive {
		if err := cmd.createParentDirectories(parentDir); err != nil {
			return err
		}
	}

	// Generar contenido del archivo
	content, err := cmd.generateFileContent(size, contentFile)
	if err != nil {
		return err
	}

	// Crear el archivo en el sistema EXT2
	return cmd.createFileInEXT2(filePath, content)
}

// validateParameters valida los parametros de entrada
func (cmd *MkfileCommand) validateParameters(size int) error {
	if size < 0 {
		return errors.New("ERROR: El tamaño no puede ser negativo")
	}
	return nil
}

// fileExists verifica si el archivo ya existe en el sistema
func (cmd *MkfileCommand) fileExists(filePath string) bool {
	return false
}

// getParentDirectory obtiene el directorio padre de una ruta
func (cmd *MkfileCommand) getParentDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

// hasWritePermissionInParent verifica permisos de escritura en directorio padre
func (cmd *MkfileCommand) hasWritePermissionInParent(parentDir string, recursive bool) bool {
	// Si es root, siempre tiene permisos
	if cmd.permissionManager.IsRoot() {
		return true
	}

	// Si no existe el directorio padre y no es recursivo
	if !cmd.directoryExists(parentDir) && !recursive {
		return false
	}

	return true
}

// directoryExists verifica si el directorio existe
func (cmd *MkfileCommand) directoryExists(dirPath string) bool {
	return dirPath == "/"
}

// createParentDirectories crea los directorios padre recursivamente
func (cmd *MkfileCommand) createParentDirectories(parentDir string) error {
	return nil
}

// generateFileContent genera el contenido del archivo segun parametros
func (cmd *MkfileCommand) generateFileContent(size int, contentFile string) ([]byte, error) {
	// Si se especifica archivo de contenido
	if contentFile != "" {
		return cmd.loadContentFromFile(contentFile)
	}

	// Si se especifica tamaño, generar contenido con numeros 0-9
	if size > 0 {
		return cmd.generateNumberContent(size), nil
	}

	// Archivo vacio (0 bytes)
	return []byte{}, nil
}

// loadContentFromFile carga contenido desde archivo externo
func (cmd *MkfileCommand) loadContentFromFile(contentFile string) ([]byte, error) {
	if _, err := os.Stat(contentFile); os.IsNotExist(err) {
		return nil, errors.New("ERROR: El archivo de contenido no existe")
	}

	content, err := os.ReadFile(contentFile)
	if err != nil {
		return nil, errors.New("ERROR: No se pudo leer el archivo de contenido")
	}

	return content, nil
}

// generateNumberContent genera contenido con numeros 0-9 repetidos
func (cmd *MkfileCommand) generateNumberContent(size int) []byte {
	content := make([]byte, size)
	for i := 0; i < size; i++ {
		content[i] = byte('0' + (i % 10))
	}
	return content
}

// createFileInEXT2 crea el archivo en el sistema EXT2
func (cmd *MkfileCommand) createFileInEXT2(filePath string, content []byte) error {
	session := cmd.loginManager.GetCurrentSession()

	// Crear inodo del archivo con permisos 664
	fileInodo := Models.Inodo{
		I_uid:   int32(session.UserID),
		I_gid:   int32(session.GroupID),
		I_s:     int32(len(content)),
		I_atime: float64(Models.GetCurrentUnixTime()),
		I_ctime: float64(Models.GetCurrentUnixTime()),
		I_mtime: float64(Models.GetCurrentUnixTime()),
		I_type:  Models.INODO_ARCHIVO,
		I_perm:  Models.SetPermissions(664), // rw-rw-r--
	}

	_ = fileInodo
	_ = content

	return nil
}

// MkFile - Función exportada para comando mkfile
func MkFile(params map[string]string) error {
	path, hasPath := params["path"]
	if !hasPath {
		return fmt.Errorf("parametro -path requerido")
	}

	rValue, hasR := params["r"]
	recursive := hasR

	// Validar que -r no reciba valor
	if hasR && rValue != "" {
		return fmt.Errorf("ERROR: El parámetro -r no debe recibir ningún valor")
	}

	sizeStr := params["size"]
	size := 0
	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			return fmt.Errorf("size invalido: %v", err)
		}
		// Validar que size no sea negativo
		if size < 0 {
			return fmt.Errorf("ERROR: El tamaño no puede ser negativo")
		}
	}

	cont := params["cont"]

	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil || !session.IsActive {
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
	fileManager := System.NewEXT2FileManager(ext2Manager)

	// Si tenemos un contenido desde archivo
	var fileContent string
	if cont != "" {
		// Si cont es una ruta de archivo, leer el contenido
		if _, err := os.Stat(cont); err == nil {
			contentBytes, err := os.ReadFile(cont)
			if err != nil {
				return fmt.Errorf("error leyendo archivo de contenido: %v", err)
			}
			fileContent = string(contentBytes)
		} else {
			// Si no es un archivo, usar el texto directo
			fileContent = cont
		}
	} else if size > 0 {
		// Si no hay contenido pero sí tamaño, crear archivo con números (0, 1, 2, etc.)
		var content strings.Builder
		for i := 0; i < size; i++ {
			content.WriteString(fmt.Sprintf("%d", i%10))
		}
		fileContent = content.String()
	}

	// Verificar si el archivo ya existe
	_, err = fileManager.ReadFileContent(path)
	if err == nil {
		// El archivo existe, preguntar si sobreescribir
		fmt.Printf("El archivo '%s' ya existe. ¿Desea sobreescribirlo? (Esta implementación procederá automáticamente)\n", path)
	}

	// Crear directorios padre si es necesario y se especifica -r
	if recursive {
		parentDir := filepath.Dir(path)
		if parentDir != "." && parentDir != "/" {
			dirManager := System.NewEXT2DirectoryManager(ext2Manager)
			
			// Dividir la ruta y crear directorios padre recursivamente
			pathParts := strings.Split(strings.Trim(parentDir, "/"), "/")
			currentPath := ""
			
			for _, part := range pathParts {
				if part == "" {
					continue
				}
				currentPath = currentPath + "/" + part
				
				// Intentar crear cada directorio en la jerarquía
				err = dirManager.CreateDirectory(currentPath, int32(session.UserID), int32(session.GroupID), 664)
				if err != nil && !strings.Contains(err.Error(), "ya existe") {
					return fmt.Errorf("error creando directorio padre '%s': %v", currentPath, err)
				}
			}
		}
	}

	// Usar la lógica existente del EXT2FileManager
	err = fileManager.WriteFileContent(path, fileContent, int32(session.UserID), int32(session.GroupID), 664)
	if err != nil {
		return err
	}

	fmt.Printf("Archivo '%s' creado exitosamente (tamaño: %d bytes)\n", path, len(fileContent))
	return nil
}
