package Reportes

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes/Graphviz"
	"MIA_2S2025_P1_202105668/Logica/System"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LsEntry representa una entrada de directorio para el reporte ls
type LsEntry struct {
	Permissions string
	Owner       string
	Group       string
	Size        int32
	Date        string
	Time        string
	Type        string
	Name        string
}

// GenerateLsReport genera un reporte de archivos y carpetas en formato JPG usando Graphviz
func GenerateLsReport(partitionID string, outputPath string, pathFileLS string) error {
	// Validar que se proporcione la ruta del directorio
	if pathFileLS == "" {
		pathFileLS = "/" // Por defecto, mostrar raíz
	}

	// Buscar la partición montada por ID
	mountedPartition := Disk.GetMountedPartitionByID(partitionID)
	if mountedPartition == nil {
		return fmt.Errorf("particion con ID '%s' no encontrada o no montada", partitionID)
	}

	// Verificar que el disco existe
	diskPath := mountedPartition.DiskPath
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Errorf("el disco '%s' no existe", diskPath)
	}

	// Obtener información de la partición y superbloque
	_, _, err := Users.GetPartitionAndSuperBlock(mountedPartition)
	if err != nil {
		return fmt.Errorf("error obteniendo superbloque: %v", err)
	}

	// Convertir MountInfo de Disk a System
	systemMountInfo := &System.MountInfo{
		DiskPath:      mountedPartition.DiskPath,
		PartitionName: mountedPartition.PartitionName,
		MountID:       mountedPartition.MountID,
		DiskLetter:    mountedPartition.DiskLetter,
		PartNumber:    mountedPartition.PartNumber,
	}

	// Crear EXT2Manager
	ext2Manager := System.NewEXT2Manager(systemMountInfo)
	if ext2Manager == nil {
		return fmt.Errorf("error creando EXT2Manager")
	}

	// Crear DirectoryManager
	dirManager := System.NewEXT2DirectoryManager(ext2Manager)
	if dirManager == nil {
		return fmt.Errorf("error creando DirectoryManager")
	}

	// Listar contenido del directorio
	entries, err := listDirectoryEntries(dirManager, pathFileLS)
	if err != nil {
		return fmt.Errorf("error listando directorio '%s': %v", pathFileLS, err)
	}

	// Crear directorio de salida si no existe
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %v", err)
	}

	// Obtener nombre del disco para mostrar en el reporte
	diskName := filepath.Base(diskPath)

	// Convertir entries a slices para Graphviz
	permissions := make([]string, len(entries))
	owners := make([]string, len(entries))
	groups := make([]string, len(entries))
	sizes := make([]int32, len(entries))
	dates := make([]string, len(entries))
	times := make([]string, len(entries))
	types := make([]string, len(entries))
	names := make([]string, len(entries))

	for i, entry := range entries {
		permissions[i] = entry.Permissions
		owners[i] = entry.Owner
		groups[i] = entry.Group
		sizes[i] = entry.Size
		dates[i] = entry.Date
		times[i] = entry.Time
		types[i] = entry.Type
		names[i] = entry.Name
	}

	// Generar el reporte usando Graphviz
	err = Graphviz.GenerateLsGraph(permissions, owners, groups, sizes, dates, times, types, names, diskName, pathFileLS, outputPath)
	if err != nil {
		return fmt.Errorf("error generando reporte ls: %v", err)
	}

	fmt.Println("Reporte ls generado exitosamente")
	return nil
}

// listDirectoryEntries lista las entradas de un directorio y convierte a LsEntry
func listDirectoryEntries(dirManager *System.EXT2DirectoryManager, dirPath string) ([]LsEntry, error) {
	// Obtener contenido del directorio
	contents, err := dirManager.ListDirectory(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error listando directorio: %v", err)
	}

	var entries []LsEntry

	for _, content := range contents {
		// Usar las fechas reales del inodo (usar MTime como fecha principal)
		date, timeStr := formatInodeTime(content.MTime)

		entry := LsEntry{
			Name:        content.Name,
			Size:        content.Size,
			Permissions: formatPermissions(content.Permissions),
			Owner:       getUserName(content.UID),
			Group:       getGroupName(content.GID),
			Date:        date,
			Time:        timeStr,
			Type:        getFileTypeString(int32(content.Type)),
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// formatInodeTime convierte timestamp Unix a formato legible
func formatInodeTime(unixTime float64) (string, string) {
	// Convertir Unix timestamp a time.Time
	timestamp := time.Unix(int64(unixTime), 0)

	// Formatear como dd/mm/yyyy y hh:mm
	date := timestamp.Format("02/01/2006")
	timeStr := timestamp.Format("15:04")

	return date, timeStr
}

// formatPermissions convierte permisos octales (ej: 755) a formato string readable
func formatPermissions(perm int32) string {
	// Extraer permisos de propietario, grupo y otros desde formato decimal (755 -> 7,5,5)
	owner := (perm / 100) % 10  // Primer dígito
	group := (perm / 10) % 10   // Segundo dígito
	other := perm % 10          // Tercer dígito

	// Convertir cada grupo de permisos a string
	ownerStr := convertPermGroup(owner)
	groupStr := convertPermGroup(group)
	otherStr := convertPermGroup(other)

	return "-" + ownerStr + groupStr + otherStr
}

// convertPermGroup convierte un valor de permisos (0-7) a string rwx
func convertPermGroup(perm int32) string {
	switch perm {
	case 0:
		return "---"
	case 1:
		return "--x"
	case 2:
		return "-w-"
	case 3:
		return "-wx"
	case 4:
		return "r--"
	case 5:
		return "r-x"
	case 6:
		return "rw-"
	case 7:
		return "rwx"
	default:
		return "---"
	}
}

// getUserName convierte UID a nombre de usuario
func getUserName(uid int32) string {
	// TODO: Implementar lookup real de usuarios desde users.txt
	// Por ahora usar nombres conocidos
	if uid == 1 {
		return "root"
	}
	return fmt.Sprintf("user%d", uid)
}

// getGroupName convierte GID a nombre de grupo
func getGroupName(gid int32) string {
	// TODO: Implementar lookup real de grupos desde users.txt
	// Por ahora usar nombres conocidos
	if gid == 1 {
		return "root"
	}
	return fmt.Sprintf("group%d", gid)
}

// getFileTypeString convierte el tipo de archivo a string
func getFileTypeString(fileType int32) string {
	switch fileType {
	case Models.INODO_ARCHIVO:
		return "Archivo"
	case Models.INODO_DIRECTORIO:
		return "Carpeta"
	default:
		return "Desconocido"
	}
}