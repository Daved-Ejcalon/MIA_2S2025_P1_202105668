package Comandos

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"fmt"
)

// MkusrCommand maneja la creacion de usuarios
type MkusrCommand struct {
	loginManager *Users.LoginManager
	userManager  *Users.UserManager
}

// NewMkusrCommand crea una nueva instancia del comando mkusr
func NewMkusrCommand(loginManager *Users.LoginManager, userManager *Users.UserManager) *MkusrCommand {
	return &MkusrCommand{
		loginManager: loginManager,
		userManager:  userManager,
	}
}

// Execute ejecuta el comando mkusr con los parametros user, pass y grp
func (cmd *MkusrCommand) Execute(username, password, groupName string) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Verificar permisos de usuario root
	session := cmd.loginManager.GetCurrentSession()
	if session.Username != "root" {
		return errors.New("ERROR: Solo el usuario root puede crear usuarios")
	}

	// Validar longitud de parametros (max 10 caracteres)
	if len(username) > 10 {
		return errors.New("ERROR: El nombre de usuario no puede exceder 10 caracteres")
	}
	if len(password) > 10 {
		return errors.New("ERROR: La contraseña no puede exceder 10 caracteres")
	}
	if len(groupName) > 10 {
		return errors.New("ERROR: El nombre del grupo no puede exceder 10 caracteres")
	}

	// Leer registros actuales
	records, err := cmd.userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Verificar que el usuario no exista (case sensitive)
	if cmd.userManager.FindUserByName(records, username) != nil {
		return errors.New("ERROR: El usuario ya existe")
	}

	// Verificar que el grupo exista (case sensitive)
	if cmd.userManager.FindGroupByName(records, groupName) == nil {
		return errors.New("ERROR: El grupo no existe")
	}

	// Crear nuevo usuario
	newUser := &Models.UserRecord{
		ID:       cmd.userManager.GetNextUserID(records),
		Type:     "U",
		Group:    groupName,
		Username: username,
		Password: password,
	}

	// Agregar y guardar
	records = append(records, newUser)
	return cmd.userManager.WriteUsersFile(records)
}

// MkUsr - Función exportada para comando mkusr
func MkUsr(params map[string]string) error {
	user, hasUser := params["user"]
	if !hasUser {
		return fmt.Errorf("parametro -user requerido")
	}

	password, hasPwd := params["pass"]
	if !hasPwd {
		return fmt.Errorf("parametro -pass requerido")
	}

	grp, hasGrp := params["grp"]
	if !hasGrp {
		return fmt.Errorf("parametro -grp requerido")
	}

	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil || !session.IsActive {
		return fmt.Errorf("ERROR: Debe iniciar sesión para usar este comando")
	}

	// Verificar permisos de usuario root
	if session.Username != "root" {
		return fmt.Errorf("ERROR: Solo el usuario root puede crear usuarios")
	}

	// Validar longitud de parametros (max 10 caracteres)
	if len(user) > 10 {
		return fmt.Errorf("ERROR: El nombre de usuario no puede exceder 10 caracteres")
	}
	if len(password) > 10 {
		return fmt.Errorf("ERROR: La contraseña no puede exceder 10 caracteres")
	}
	if len(grp) > 10 {
		return fmt.Errorf("ERROR: El nombre del grupo no puede exceder 10 caracteres")
	}

	// Obtener UserManager para la sesión activa
	mountInfo, err := Disk.GetMountInfoByID(session.MountID)
	if err != nil {
		return fmt.Errorf("partición no encontrada: %v", err)
	}

	partitionInfo, superBloque, err := Users.GetPartitionAndSuperBlock(mountInfo)
	if err != nil {
		return fmt.Errorf("error accediendo al sistema de archivos: %v", err)
	}

	userManager := Users.NewUserManager(mountInfo.DiskPath, partitionInfo, superBloque)

	// Usar la lógica existente del UserManager
	err = userManager.CreateUser(user, grp, password)
	if err != nil {
		return err
	}

	fmt.Printf("User: \"%s\" creado en el grupo \"%s\"\n", user, grp)
	return nil
}
