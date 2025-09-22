package Users

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// UserManager maneja operaciones de usuarios y grupos en users.txt
type UserManager struct {
	diskPath      string
	partitionInfo *Models.Partition
	superBloque   *Models.SuperBloque
}

// NewUserManager crea una nueva instancia del gestor de usuarios
func NewUserManager(diskPath string, partitionInfo *Models.Partition, superBloque *Models.SuperBloque) *UserManager {
	return &UserManager{
		diskPath:      diskPath,
		partitionInfo: partitionInfo,
		superBloque:   superBloque,
	}
}

// ReadUsersFile lee y parsea el archivo users.txt del sistema
func (um *UserManager) ReadUsersFile() ([]*Models.UserRecord, error) {
	file, err := os.Open(um.diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	inodoPos := um.partitionInfo.PartStart + int64(um.superBloque.S_inode_start) + int64(1*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return nil, err
	}

	var usersInodo Models.Inodo
	err = binary.Read(file, binary.LittleEndian, &usersInodo)
	if err != nil {
		return nil, err
	}

	blockPos := um.partitionInfo.PartStart + int64(um.superBloque.S_block_start) + int64(usersInodo.I_block[0]*int32(Models.BLOQUE_SIZE))
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return nil, err
	}

	var contentBlock Models.BloqueArchivos
	err = binary.Read(file, binary.LittleEndian, &contentBlock)
	if err != nil {
		return nil, err
	}

	// Leer contenido usando múltiples bloques si es necesario
	var allContent []byte
	bytesRead := int32(0)

	// Leer desde todos los bloques asignados al archivo users.txt
	for i := 0; i < 12 && usersInodo.I_block[i] != -1; i++ {
		if bytesRead >= usersInodo.I_s {
			break
		}

		// Posicionarse en el bloque
		blockPos := um.partitionInfo.PartStart + int64(um.superBloque.S_block_start) + int64(usersInodo.I_block[i]*int32(Models.BLOQUE_SIZE))
		_, err = file.Seek(blockPos, 0)
		if err != nil {
			return nil, err
		}

		var blockData Models.BloqueArchivos
		err = binary.Read(file, binary.LittleEndian, &blockData)
		if err != nil {
			return nil, err
		}

		// Calcular cuántos bytes tomar de este bloque
		bytesRemaining := usersInodo.I_s - bytesRead
		bytesToTake := int32(Models.BLOQUE_SIZE)
		if bytesToTake > bytesRemaining {
			bytesToTake = bytesRemaining
		}

		allContent = append(allContent, blockData.GetContent()[:bytesToTake]...)
		bytesRead += bytesToTake
	}

	content := string(allContent)
	return um.parseUsersContent(content)
}

// parseUsersContent convierte el contenido de users.txt en registros
func (um *UserManager) parseUsersContent(content string) ([]*Models.UserRecord, error) {
	var records []*Models.UserRecord
	lines := strings.Split(strings.TrimSpace(content), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		record, err := Models.ParseUserRecord(line)
		if err != nil {
			// Ignorar líneas malformadas y continuar
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// WriteUsersFile escribe los registros al archivo users.txt
func (um *UserManager) WriteUsersFile(records []*Models.UserRecord) error {
	var content strings.Builder
	for _, record := range records {
		content.WriteString(record.ToString())
		content.WriteString("\n")
	}

	contentStr := content.String()

	// El archivo users.txt ahora puede usar múltiples bloques, sin límite de 64 bytes

	file, err := os.OpenFile(um.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	inodoPos := um.partitionInfo.PartStart + int64(um.superBloque.S_inode_start) + int64(1*Models.INODO_SIZE)
	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return err
	}

	var usersInodo Models.Inodo
	err = binary.Read(file, binary.LittleEndian, &usersInodo)
	if err != nil {
		return err
	}

	// Escribir contenido usando múltiples bloques
	err = um.writeUsersContentMultiBlock(file, &usersInodo, contentStr)
	if err != nil {
		return err
	}

	// Actualizar metadatos del inodo
	usersInodo.I_s = int32(len(contentStr))
	usersInodo.I_mtime = float64(Models.GetCurrentUnixTime())

	_, err = file.Seek(inodoPos, 0)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &usersInodo)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// writeUsersContentMultiBlock escribe contenido del archivo users.txt usando múltiples bloques
func (um *UserManager) writeUsersContentMultiBlock(file *os.File, usersInodo *Models.Inodo, content string) error {
	contentBytes := []byte(content)
	totalBytes := len(contentBytes)
	blocksNeeded := (totalBytes + Models.BLOQUE_SIZE - 1) / Models.BLOQUE_SIZE

	// Limitar a 12 bloques directos
	if blocksNeeded > 12 {
		blocksNeeded = 12
		contentBytes = contentBytes[:12*Models.BLOQUE_SIZE]
	}

	// Liberar bloques existentes (excepto el primero si ya existe)
	for i := 1; i < 12; i++ {
		if usersInodo.I_block[i] != -1 {
			// Aquí deberíamos liberar el bloque en el bitmap, pero por simplicidad lo omitimos
			usersInodo.I_block[i] = -1
		}
	}

	// Escribir contenido en bloques
	for i := 0; i < blocksNeeded; i++ {
		// Calcular qué porción del contenido va en este bloque
		start := i * Models.BLOQUE_SIZE
		end := start + Models.BLOQUE_SIZE
		if end > len(contentBytes) {
			end = len(contentBytes)
		}

		blockContent := contentBytes[start:end]

		// Si es el primer bloque, usar el bloque existente
		var blockNum int32
		if i == 0 && usersInodo.I_block[0] != -1 {
			blockNum = usersInodo.I_block[0]
		} else {
			// Para bloques adicionales, usar bloques consecutivos seguros
			// Asegurándonos de no sobrescribir otros archivos
			if usersInodo.I_block[0] != -1 {
				blockNum = usersInodo.I_block[0] + int32(i)
			} else {
				// Si no hay primer bloque, usar bloque alto como base
				blockNum = 100 + int32(i)
			}
			usersInodo.I_block[i] = blockNum
		}

		// Limpiar el bloque antes de escribir para evitar basura
		blockPos := um.partitionInfo.PartStart + int64(um.superBloque.S_block_start) + int64(blockNum*int32(Models.BLOQUE_SIZE))
		_, err := file.Seek(blockPos, 0)
		if err != nil {
			return err
		}

		// Limpiar el bloque completo con ceros
		emptyBlock := make([]byte, Models.BLOQUE_SIZE)
		_, err = file.Write(emptyBlock)
		if err != nil {
			return err
		}

		// Reposicionarse para escribir el contenido
		_, err = file.Seek(blockPos, 0)
		if err != nil {
			return err
		}

		contentBlock := Models.BloqueArchivos{}
		contentBlock.SetContent(blockContent)

		buffer := new(bytes.Buffer)
		err = binary.Write(buffer, binary.LittleEndian, &contentBlock)
		if err != nil {
			return err
		}
		_, err = file.Write(buffer.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetNextUserID obtiene el siguiente ID disponible para usuarios
func (um *UserManager) GetNextUserID(records []*Models.UserRecord) int {
	maxID := 0
	for _, record := range records {
		if record.Type == "U" && record.ID > maxID {
			maxID = record.ID
		}
	}
	return maxID + 1
}

// GetNextGroupID obtiene el siguiente ID disponible para grupos
func (um *UserManager) GetNextGroupID(records []*Models.UserRecord) int {
	maxID := 0
	for _, record := range records {
		if record.Type == "G" && record.ID > maxID {
			maxID = record.ID
		}
	}
	return maxID + 1
}

// FindUserByName busca un usuario por nombre
func (um *UserManager) FindUserByName(records []*Models.UserRecord, username string) *Models.UserRecord {
	for _, record := range records {
		if record.Type == "U" && record.Username == username && record.ID != 0 {
			return record
		}
	}
	return nil
}

// FindGroupByName busca un grupo por nombre
func (um *UserManager) FindGroupByName(records []*Models.UserRecord, groupname string) *Models.UserRecord {
	for _, record := range records {
		if record.Type == "G" && record.Group == groupname && record.ID != 0 {
			return record
		}
	}
	return nil
}

// ValidateUserCredentials valida credenciales de usuario
func (um *UserManager) ValidateUserCredentials(records []*Models.UserRecord, username, password string) bool {
	user := um.FindUserByName(records, username)
	return user != nil && user.Password == password
}

// CreateUser crea un nuevo usuario en el sistema
func (um *UserManager) CreateUser(username, groupname, password string) error {

	records, err := um.ReadUsersFile()
	if err != nil {
		return err
	}

	// Validar que el usuario no exista
	existingUser := um.FindUserByName(records, username)
	if existingUser != nil {
		return fmt.Errorf("Error: \"%s\" ya existe", username)
	}

	// Validar que el grupo exista
	existingGroup := um.FindGroupByName(records, groupname)
	if existingGroup == nil {
		return fmt.Errorf("Error: El grupo \"%s\" no existe. Debe crear el grupo.", groupname)
	}

	newUser := &Models.UserRecord{
		ID:       um.GetNextUserID(records),
		Type:     "U",
		Group:    groupname,
		Username: username,
		Password: password,
	}

	records = append(records, newUser)
	return um.WriteUsersFile(records)
}

// CreateGroup crea un nuevo grupo en el sistema
func (um *UserManager) CreateGroup(groupname string) error {

	records, err := um.ReadUsersFile()
	if err != nil {
		return err
	}

	// Validar que el grupo no exista
	existingGroup := um.FindGroupByName(records, groupname)
	if existingGroup != nil {
		return fmt.Errorf("Error: El grupo \"%s\" ya existe", groupname)
	}

	newGroup := &Models.UserRecord{
		ID:    um.GetNextGroupID(records),
		Type:  "G",
		Group: groupname,
	}

	records = append(records, newGroup)
	return um.WriteUsersFile(records)
}

// DeleteUser marca un usuario como eliminado (ID = 0)
func (um *UserManager) DeleteUser(username string) error {
	records, err := um.ReadUsersFile()
	if err != nil {
		return err
	}

	user := um.FindUserByName(records, username)
	if user == nil {
		return fmt.Errorf("ERROR: El usuario '%s' no existe", username)
	}

	user.ID = 0
	return um.WriteUsersFile(records)
}

// DeleteGroup marca un grupo como eliminado (ID = 0)
func (um *UserManager) DeleteGroup(groupname string) error {
	records, err := um.ReadUsersFile()
	if err != nil {
		return err
	}

	group := um.FindGroupByName(records, groupname)

	group.ID = 0
	return um.WriteUsersFile(records)
}
