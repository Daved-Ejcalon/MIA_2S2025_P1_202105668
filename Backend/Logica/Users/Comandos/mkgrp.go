package Comandos

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"errors"
	"fmt"
)

// MkgrpCommand maneja la creacion de grupos
type MkgrpCommand struct {
	loginManager *Users.LoginManager
	userManager  *Users.UserManager
}

// NewMkgrpCommand crea una nueva instancia del comando mkgrp
func NewMkgrpCommand(loginManager *Users.LoginManager, userManager *Users.UserManager) *MkgrpCommand {
	return &MkgrpCommand{
		loginManager: loginManager,
		userManager:  userManager,
	}
}

// Execute ejecuta el comando mkgrp con el parametro name
func (cmd *MkgrpCommand) Execute(groupName string) error {
	// Verificar sesion activa
	if err := cmd.loginManager.RequireSession(); err != nil {
		return err
	}

	// Verificar permisos de usuario root
	session := cmd.loginManager.GetCurrentSession()
	if session.Username != "root" {
		return errors.New("ERROR: Solo el usuario root puede crear grupos")
	}

	// Leer registros actuales
	records, err := cmd.userManager.ReadUsersFile()
	if err != nil {
		return err
	}

	// Verificar que el grupo no exista (case sensitive)
	if cmd.userManager.FindGroupByName(records, groupName) != nil {
		return errors.New("ERROR: El grupo ya existe")
	}

	// Crear nuevo grupo
	newGroup := &Models.UserRecord{
		ID:    cmd.userManager.GetNextGroupID(records),
		Type:  "G",
		Group: groupName,
	}

	// Agregar y guardar
	records = append(records, newGroup)
	return cmd.userManager.WriteUsersFile(records)
}

// MkGrp - Función exportada para comando mkgrp
func MkGrp(params map[string]string) error {
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
		return fmt.Errorf("ERROR: Solo el usuario root puede crear grupos")
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
	err = userManager.CreateGroup(name)
	if err != nil {
		return err
	}

	return nil
}
