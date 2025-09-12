package Users

import (
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
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

	blockPos := um.partitionInfo.PartStart + int64(um.superBloque.S_block_start) + int64(usersInodo.I_block[0]*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return nil, err
	}

	var contentBlock Models.BloqueArchivos
	err = binary.Read(file, binary.LittleEndian, &contentBlock)
	if err != nil {
		return nil, err
	}

	content := string(contentBlock.GetContent()[:usersInodo.I_s])
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

		record, _ := Models.ParseUserRecord(line)

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

	blockPos := um.partitionInfo.PartStart + int64(um.superBloque.S_block_start) + int64(usersInodo.I_block[0]*Models.BLOQUE_SIZE)
	_, err = file.Seek(blockPos, 0)
	if err != nil {
		return err
	}

	contentBlock := Models.BloqueArchivos{}
	contentBlock.SetContent([]byte(contentStr))

	buffer = new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, &contentBlock)
	if err != nil {
		return err
	}
	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
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
