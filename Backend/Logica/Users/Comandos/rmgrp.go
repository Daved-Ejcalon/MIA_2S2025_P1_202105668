package Comandos

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"errors"
	"fmt"
)

// RmgrpCommand maneja la eliminacion de grupos
type RmgrpCommand struct {
	loginManager *Users.LoginManager
	userManager  *Users.UserManager
}

// NewRmgrpCommand crea una nueva instancia del comando rmgrp
func NewRmgrpCommand(loginManager *Users.LoginManager, userManager *Users.UserManager) *RmgrpCommand {
	return &RmgrpCommand{
		loginManager: loginManager,
		userManager:  userManager,
	}
}

// Execute ejecuta el comando rmgrp con el parametro name
func (cmd *RmgrpCommand) Execute(groupName string) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Verificar permisos de usuario root
	session := cmd.loginManager.GetCurrentSession()
	if session.Username != "root" {
		return errors.New("ERROR: Solo el usuario root puede eliminar grupos")
	}

	// Leer registros actuales
	records, err := cmd.userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Buscar el grupo a eliminar (case sensitive)
	group := cmd.userManager.FindGroupByName(records, groupName)
	if group == nil {
		return errors.New("ERROR: El grupo no existe")
	}

	// Marcar grupo como eliminado (ID = 0)
	group.ID = 0

	// Guardar cambios
	return cmd.userManager.WriteUsersFile(records)
}

// RmGrp - Función exportada para comando rmgrp
func RmGrp(params map[string]string) error {
	name, hasName := params["name"]
	if !hasName {
		return fmt.Errorf("parametro -name requerido")
	}

	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil || !session.IsActive {
		return fmt.Errorf("ERROR: Debe iniciar sesión para usar este comando")
	}

	// Verificar permisos de usuario root
	if session.Username != "root" {
		return fmt.Errorf("ERROR: Solo el usuario root puede eliminar grupos")
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
	err = userManager.DeleteGroup(name)
	if err != nil {
		return err
	}

	fmt.Printf("Grupo '%s' eliminado exitosamente\n", name)
	return nil
}
