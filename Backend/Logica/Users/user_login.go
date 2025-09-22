package Users

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// Session representa la sesion activa actual
type Session struct {
	IsActive bool
	Username string
	UserID   int
	GroupID  int
	MountID  string
}

// LoginManager maneja las operaciones de login y logout
type LoginManager struct {
	currentSession *Session
}

// NewLoginManager crea una nueva instancia del gestor de login
func NewLoginManager() *LoginManager {
	return &LoginManager{
		currentSession: &Session{
			IsActive: false,
		},
	}
}

// IsLoggedIn verifica si hay una sesion activa
func (lm *LoginManager) IsLoggedIn() bool {
	return lm.currentSession.IsActive
}

// GetCurrentSession obtiene la sesion actual
func (lm *LoginManager) GetCurrentSession() *Session {
	return lm.currentSession
}

// ValidateMountID verifica si el ID de particion montada existe
func (lm *LoginManager) ValidateMountID(mountID string) bool {
	// Esta funcion deberia verificar contra la lista de particiones montadas
	// Por ahora asumimos que el ID es valido si no esta vacio
	return mountID != ""
}

// Login autentica un usuario con los parametros del comando
func (lm *LoginManager) Login(username, password, mountID string, userManager *UserManager) error {
	// Verificar si ya hay una sesion activa
	if lm.IsLoggedIn() {
		return errors.New("Error: Sesión activa")
	}

	// Validar que el mount ID exista
	if !lm.ValidateMountID(mountID) {
		return fmt.Errorf("ERROR: La partición con ID '%s' no está montada", mountID)
	}

	records, err := userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Buscar usuario (case sensitive)
	user := userManager.FindUserByName(records, username)
	if user == nil {
		return fmt.Errorf("ERROR: El usuario '%s' no existe", username)
	}

	// Validar contraseña (case sensitive)
	if user.Password != password {
		return errors.New("ERROR: Contraseña incorrecta")
	}

	// Obtener grupo del usuario
	group := userManager.FindGroupByName(records, user.Group)

	lm.currentSession = &Session{
		IsActive: true,
		Username: username,
		UserID:   user.ID,
		GroupID:  group.ID,
		MountID:  mountID,
	}

	return nil
}

// Logout cierra la sesion actual
func (lm *LoginManager) Logout() error {
	// Verificar que haya una sesion activa
	if !lm.IsLoggedIn() {
		return errors.New("ERROR: No hay ninguna sesión activa")
	}

	lm.currentSession = &Session{
		IsActive: false,
	}

	return nil
}

// RequireSession verifica que haya una sesion activa para comandos que la necesitan
func (lm *LoginManager) RequireSession() error {
	if !lm.IsLoggedIn() {
		return errors.New("ERROR: Debe iniciar sesión para usar este comando")
	}
	return nil
}

// Variable global para manejar las sesiones
var loginManager = NewLoginManager()

// GetCurrentSession obtiene la sesión actual (función exportada)
func GetCurrentSession() *Session {
	return loginManager.GetCurrentSession()
}

// Login - Función exportada para comando login
func Login(params map[string]string) error {
	username, hasUsername := params["user"]
	if !hasUsername {
		return fmt.Errorf("parametro -user requerido")
	}

	password, hasPassword := params["pass"]
	if !hasPassword {
		return fmt.Errorf("parametro -pass requerido")
	}

	mountID, hasID := params["id"]
	if !hasID {
		return fmt.Errorf("parametro -id requerido")
	}

	// Verificar si ya hay una sesión activa
	if loginManager.IsLoggedIn() {
		return errors.New("Error: Sesión activa")
	}

	// Verificar que el mount ID exista
	mountInfo, err := Disk.GetMountInfoByID(mountID)
	if err != nil {
		return fmt.Errorf("ERROR: La partición con ID '%s' no está montada", mountID)
	}

	// Obtener información de la partición y leer users.txt
	partitionInfo, superBloque, err := GetPartitionAndSuperBlock(mountInfo)
	if err != nil {
		return fmt.Errorf("ERROR: No se pudo acceder al sistema de archivos: %v", err)
	}

	// Crear UserManager para esta partición
	userManager := NewUserManager(mountInfo.DiskPath, partitionInfo, superBloque)
	
	// Leer archivo users.txt
	records, err := userManager.ReadUsersFile()
	if err != nil {
		return fmt.Errorf("ERROR: No se pudo leer el archivo de usuarios: %v", err)
	}

	// Buscar usuario (case sensitive)
	user := userManager.FindUserByName(records, username)
	if user == nil {
		return fmt.Errorf("ERROR: El usuario '%s' no existe", username)
	}

	// Validar contraseña (case sensitive)
	if user.Password != password {
		return errors.New("ERROR: Contraseña incorrecta")
	}

	// Obtener grupo del usuario
	group := userManager.FindGroupByName(records, user.Group)
	if group == nil {
		return fmt.Errorf("ERROR: El grupo '%s' del usuario no existe", user.Group)
	}

	// Crear sesión exitosa
	loginManager.currentSession = &Session{
		IsActive: true,
		Username: username,
		UserID:   user.ID,
		GroupID:  group.ID,
		MountID:  mountID,
	}

	fmt.Printf("Login %s: id=%s\n", username, mountID)
	return nil
}

// GetPartitionAndSuperBlock obtiene información de partición y SuperBloque (función exportada)
func GetPartitionAndSuperBlock(mountInfo *Disk.MountInfo) (*Models.Partition, *Models.SuperBloque, error) {
	file, err := os.Open(mountInfo.DiskPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return nil, nil, err
	}

	// Buscar la partición montada
	var targetPartition *Models.Partition
	for i, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == mountInfo.PartitionName {
			targetPartition = &mbr.Partitions[i]
			break
		}
	}

	if targetPartition == nil {
		return nil, nil, fmt.Errorf("partición no encontrada")
	}

	// Leer SuperBloque desde el inicio de la partición
	superBloquePos := targetPartition.PartStart
	file.Seek(superBloquePos, 0)
	
	var superBloque Models.SuperBloque
	err = binary.Read(file, binary.LittleEndian, &superBloque)
	if err != nil {
		return nil, nil, fmt.Errorf("error leyendo SuperBloque: %v", err)
	}

	return targetPartition, &superBloque, nil
}

// Logout - Función exportada para comando logout
func Logout() error {
	return loginManager.Logout()
}
