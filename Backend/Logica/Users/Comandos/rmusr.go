package Comandos

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"errors"
	"fmt"
)

// RmusrCommand maneja la eliminacion de usuarios
type RmusrCommand struct {
	loginManager *Users.LoginManager
	userManager  *Users.UserManager
}

// NewRmusrCommand crea una nueva instancia del comando rmusr
func NewRmusrCommand(loginManager *Users.LoginManager, userManager *Users.UserManager) *RmusrCommand {
	return &RmusrCommand{
		loginManager: loginManager,
		userManager:  userManager,
	}
}

// Execute ejecuta el comando rmusr con el parametro user
func (cmd *RmusrCommand) Execute(username string) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Verificar permisos de usuario root
	session := cmd.loginManager.GetCurrentSession()
	if session.Username != "root" {
		return errors.New("ERROR: Solo el usuario root puede eliminar usuarios")
	}

	// Leer registros actuales
	records, err := cmd.userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Buscar el usuario a eliminar (case sensitive)
	user := cmd.userManager.FindUserByName(records, username)
	if user == nil {
		return errors.New("ERROR: El usuario no existe")
	}

	// Marcar usuario como eliminado (ID = 0)
	user.ID = 0

	// Guardar cambios
	return cmd.userManager.WriteUsersFile(records)
}

// RmUsr - Función exportada para comando rmusr
func RmUsr(params map[string]string) error {
	usr, hasUsr := params["user"]
	if !hasUsr {
		return fmt.Errorf("parametro -user requerido")
	}

	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil || !session.IsActive {
		return fmt.Errorf("ERROR: Debe iniciar sesión para usar este comando")
	}

	// Verificar permisos de usuario root
	if session.Username != "root" {
		return fmt.Errorf("ERROR: Solo el usuario root puede eliminar usuarios")
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
	err = userManager.DeleteUser(usr)
	if err != nil {
		return err
	}

	fmt.Printf("Usuario '%s' eliminado exitosamente\n", usr)
	return nil
}
