package Models

import (
	"fmt"
	"strconv"
	"strings"
)

// UserRecord representa un registro de usuario o grupo en users.txt
type UserRecord struct {
	ID       int
	Type     string // "U" para Usuario, "G" para Grupo
	Group    string
	Username string // Solo para usuarios
	Password string // Solo para usuarios
}

// ToString convierte el registro a formato string para users.txt
func (u *UserRecord) ToString() string {
	if u.Type == "U" {
		return fmt.Sprintf("%d, U, %s, %s, %s", u.ID, u.Group, u.Username, u.Password)
	}
	return fmt.Sprintf("%d, G, %s", u.ID, u.Group)
}

// ParseUserRecord convierte una linea de users.txt a UserRecord
func ParseUserRecord(line string) (*UserRecord, error) {
	parts := strings.Split(strings.TrimSpace(line), ",")

	// Validar que la línea tenga al menos 3 partes (ID, Type, Group)
	if len(parts) < 3 {
		return nil, fmt.Errorf("línea malformada: %s", line)
	}

	// Limpiar espacios de cada parte
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	id, _ := strconv.Atoi(parts[0])
	record := &UserRecord{
		ID:   id,
		Type: parts[1],
	}

	if parts[1] == "G" {
		record.Group = parts[2]
	} else {
		record.Group = parts[2]
		record.Username = parts[3]
		record.Password = parts[4]
	}

	return record, nil
}

// CreateInitialUsersContent genera el contenido inicial de users.txt
func CreateInitialUsersContent() string {
	return "1, G, root\n1, U, root, root, 123\n"
}
