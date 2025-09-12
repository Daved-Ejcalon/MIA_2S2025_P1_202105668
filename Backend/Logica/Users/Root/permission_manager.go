package Root

import (
	"MIA_2S2025_P1_202105668/Logica/Users"
)

// UserCategory representa la categoria del usuario respecto al archivo
type UserCategory int

const (
	USER_OWNER   UserCategory = iota // Usuario propietario (U)
	GROUP_MEMBER                     // Usuario del mismo grupo (G)
	OTHER_USER                       // Otro usuario (O)
)

// PermissionManager maneja la logica de permisos UGO
type PermissionManager struct {
	loginManager *Users.LoginManager
}

// NewPermissionManager crea una nueva instancia del gestor de permisos
func NewPermissionManager(loginManager *Users.LoginManager) *PermissionManager {
	return &PermissionManager{
		loginManager: loginManager,
	}
}

// IsRoot verifica si el usuario actual es root
func (pm *PermissionManager) IsRoot() bool {
	session := pm.loginManager.GetCurrentSession()
	return session.IsActive && session.Username == "root"
}

// GetUserCategory determina la categoria del usuario actual respecto a un archivo/carpeta
func (pm *PermissionManager) GetUserCategory(fileOwnerID, fileGroupID int) UserCategory {
	session := pm.loginManager.GetCurrentSession()

	// Si es el propietario del archivo
	if session.UserID == fileOwnerID {
		return USER_OWNER
	}

	// Si pertenece al mismo grupo del archivo
	if session.GroupID == fileGroupID {
		return GROUP_MEMBER
	}

	// Otro usuario
	return OTHER_USER
}

// HasPermission verifica si el usuario actual tiene permisos sobre un archivo/carpeta
func (pm *PermissionManager) HasPermission(fileOwnerID, fileGroupID int, filePermissions int32, requiredPermission int) bool {
	// Root siempre tiene permisos 777
	if pm.IsRoot() {
		return true
	}

	category := pm.GetUserCategory(fileOwnerID, fileGroupID)

	// Extraer permisos segun categoria UGO
	var userPerms int
	switch category {
	case USER_OWNER:
		// Permisos del propietario (bits 8-6)
		userPerms = int((filePermissions >> 6) & 7)
	case GROUP_MEMBER:
		// Permisos del grupo (bits 5-3)
		userPerms = int((filePermissions >> 3) & 7)
	case OTHER_USER:
		// Permisos de otros (bits 2-0)
		userPerms = int(filePermissions & 7)
	}

	// Verificar si tiene el permiso requerido
	return (userPerms & requiredPermission) != 0
}

// HasReadPermission verifica permiso de lectura (r=4)
func (pm *PermissionManager) HasReadPermission(fileOwnerID, fileGroupID int, filePermissions int32) bool {
	return pm.HasPermission(fileOwnerID, fileGroupID, filePermissions, 4)
}

// HasWritePermission verifica permiso de escritura (w=2)
func (pm *PermissionManager) HasWritePermission(fileOwnerID, fileGroupID int, filePermissions int32) bool {
	return pm.HasPermission(fileOwnerID, fileGroupID, filePermissions, 2)
}

// HasExecutePermission verifica permiso de ejecucion (x=1)
func (pm *PermissionManager) HasExecutePermission(fileOwnerID, fileGroupID int, filePermissions int32) bool {
	return pm.HasPermission(fileOwnerID, fileGroupID, filePermissions, 1)
}
