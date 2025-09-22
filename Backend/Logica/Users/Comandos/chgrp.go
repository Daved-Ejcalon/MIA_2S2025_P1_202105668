package Comandos

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"errors"
	"fmt"
)

// ChgrpCommand maneja el cambio de grupo de usuarios
type ChgrpCommand struct {
	loginManager *Users.LoginManager
	userManager  *Users.UserManager
}

// NewChgrpCommand crea una nueva instancia del comando chgrp
func NewChgrpCommand(loginManager *Users.LoginManager, userManager *Users.UserManager) *ChgrpCommand {
	return &ChgrpCommand{
		loginManager: loginManager,
		userManager:  userManager,
	}
}

// Execute ejecuta el comando chgrp con los parametros user y grp
func (cmd *ChgrpCommand) Execute(username, newGroupName string) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Verificar permisos de usuario root
	session := cmd.loginManager.GetCurrentSession()
	if session.Username != "root" {
		return errors.New("ERROR: Solo el usuario root puede cambiar grupos")
	}

	// Leer registros actuales
	records, err := cmd.userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Buscar el usuario a modificar (case sensitive)
	user := cmd.userManager.FindUserByName(records, username)
	if user == nil {
		return errors.New("ERROR: El usuario no existe")
	}

	// Verificar que el nuevo grupo exista (case sensitive)
	if cmd.userManager.FindGroupByName(records, newGroupName) == nil {
		return errors.New("ERROR: El grupo no existe")
	}

	// Cambiar el grupo del usuario
	user.Group = newGroupName

	// Guardar cambios
	return cmd.userManager.WriteUsersFile(records)
}

// ChGrp - Función exportada para comando chgrp
func ChGrp(params map[string]string) error {
	usr, hasUsr := params["user"]
	if !hasUsr {
		return fmt.Errorf("parametro -user requerido")
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
		return fmt.Errorf("ERROR: Solo el usuario root puede cambiar grupos")
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

	// Leer registros actuales
	records, err := userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Buscar el usuario a modificar (case sensitive)
	user := userManager.FindUserByName(records, usr)
	if user == nil {
		return fmt.Errorf("ERROR: El usuario '%s' no existe", usr)
	}

	// Verificar que el nuevo grupo exista (case sensitive)
	if userManager.FindGroupByName(records, grp) == nil {
		return fmt.Errorf("ERROR: El grupo '%s' no existe", grp)
	}

	// Cambiar el grupo del usuario
	user.Group = grp

	// Guardar cambios
	err = userManager.WriteUsersFile(records)
	if err != nil {
		return err
	}

	fmt.Printf("Grupo de usuario '%s' cambiado a '%s' exitosamente\n", usr, grp)
	return nil
}
