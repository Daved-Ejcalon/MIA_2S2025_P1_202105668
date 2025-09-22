package System

import (
	"MIA_2S2025_P1_202105668/Models"
)

// ValidateFileReadPermission valida permisos de lectura para un archivo
func ValidateFileReadPermission(fileOwnerID, fileGroupID int32, permissions [3]byte, userID, userGroupID int) bool {
	// Si es root (UserID = 1), siempre tiene permisos
	if userID == 1 {
		return true
	}

	perms := Models.GetPermissions(permissions)

	// Determinar categoría del usuario
	if int32(userID) == fileOwnerID {
		// Usuario propietario - verificar bit de lectura del propietario (r--)
		return (perms & 0400) != 0 // 0400 = 100 000 000 (r-- --- ---)
	}

	if int32(userGroupID) == fileGroupID {
		// Usuario del mismo grupo - verificar bit de lectura del grupo (-r-)
		return (perms & 0040) != 0 // 0040 = 000 100 000 (--- r-- ---)
	}

	// Otro usuario - verificar bit de lectura de otros (--r)
	return (perms & 0004) != 0 // 0004 = 000 000 100 (--- --- r--)
}

// ValidateFileWritePermission valida permisos de escritura para un archivo/directorio
func ValidateFileWritePermission(fileOwnerID, fileGroupID int32, permissions [3]byte, userID, userGroupID int) bool {
	// Si es root (UserID = 1), siempre tiene permisos
	if userID == 1 {
		return true
	}

	perms := Models.GetPermissions(permissions)

	// Determinar categoría del usuario
	if int32(userID) == fileOwnerID {
		// Usuario propietario - verificar bit de escritura del propietario (-w-)
		return (perms & 0200) != 0 // 0200 = 010 000 000 (-w- --- ---)
	}

	if int32(userGroupID) == fileGroupID {
		// Usuario del mismo grupo - verificar bit de escritura del grupo (-w-)
		return (perms & 0020) != 0 // 0020 = 000 010 000 (--- -w- ---)
	}

	// Otro usuario - verificar bit de escritura de otros (-w)
	return (perms & 0002) != 0 // 0002 = 000 000 010 (--- --- -w-)
}

// ValidateFileExecutePermission valida permisos de ejecución para un archivo
func ValidateFileExecutePermission(fileOwnerID, fileGroupID int32, permissions [3]byte, userID, userGroupID int) bool {
	// Si es root (UserID = 1), siempre tiene permisos
	if userID == 1 {
		return true
	}

	perms := Models.GetPermissions(permissions)

	// Determinar categoría del usuario
	if int32(userID) == fileOwnerID {
		// Usuario propietario - verificar bit de ejecución del propietario (--x)
		return (perms & 0100) != 0 // 0100 = 001 000 000 (--x --- ---)
	}

	if int32(userGroupID) == fileGroupID {
		// Usuario del mismo grupo - verificar bit de ejecución del grupo (--x)
		return (perms & 0010) != 0 // 0010 = 000 001 000 (--- --x ---)
	}

	// Otro usuario - verificar bit de ejecución de otros (--x)
	return (perms & 0001) != 0 // 0001 = 000 000 001 (--- --- --x)
}