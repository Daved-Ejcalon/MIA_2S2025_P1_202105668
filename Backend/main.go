package main

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Logica/Users/Comandos"
	"MIA_2S2025_P1_202105668/Logica/Users/Root"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("MIA> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Ignorar líneas que empiecen con # (comentarios)
		if strings.HasPrefix(input, "#") || input == "" {
			continue
		}

		// Remover comentarios inline (después del comando)
		if commentIndex := strings.Index(input, "#"); commentIndex != -1 {
			input = strings.TrimSpace(input[:commentIndex])
		}

		// Si después de remover comentarios queda vacío, continuar
		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("saliendo del sistema...")
			break
		}

		err := processCommand(input)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
	}
}

func processCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return fmt.Errorf("comando vacio")
	}

	command := strings.ToLower(parts[0])
	params := parseParameters(parts[1:])

	switch command {
	case "mkdisk":
		return processMkdisk(params)
	case "rmdisk":
		return processRmdisk(params)
	case "fdisk":
		return processFdisk(params)
	case "mount":
		return processMount(params)
	case "mounted":
		Disk.Mounted()
		return nil
	case "mkfs":
		return processMkfs(params)
	case "cat":
		return Disk.Cat(params)
	case "showdisk":
		return Disk.ShowDisk(params)
	case "login":
		return Users.Login(params)
	case "logout":
		return Users.Logout()
	case "mkgrp":
		return Comandos.MkGrp(params)
	case "rmgrp":
		return Comandos.RmGrp(params)
	case "mkusr":
		return Comandos.MkUsr(params)
	case "rmusr":
		return Comandos.RmUsr(params)
	case "chgrp":
		return Comandos.ChGrp(params)
	case "mkdir":
		return Root.MkDir(params)
	case "mkfile":
		return Root.MkFile(params)
	default:
		return fmt.Errorf("comando '%s' no reconocido", command)
	}
}

func processMkdisk(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"size": true,
		"unit": true,
		"fit":  true,
		"path": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para mkdisk", param)
		}
	}

	sizeStr, hasSize := params["size"]
	if !hasSize {
		return fmt.Errorf("parametro -size requerido")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("size invalido: %v", err)
	}

	unit := params["unit"]
	if unit == "" {
		unit = "K"
	}

	fit := params["fit"]
	if fit == "" {
		fit = "WF"
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	return Disk.MkDisk(size, unit, fit, path)
}

func processRmdisk(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"path": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para rmdisk", param)
		}
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	return Disk.RmDisk(path)
}

func processFdisk(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"size": true,
		"unit": true,
		"fit":  true,
		"path": true,
		"type": true,
		"name": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para fdisk", param)
		}
	}

	sizeStr, hasSize := params["size"]
	if !hasSize {
		return fmt.Errorf("parametro -size requerido")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("size invalido: %v", err)
	}

	unit := params["unit"]
	if unit == "" {
		unit = "K"
	}

	fit := params["fit"]
	if fit == "" {
		fit = "WF"
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	ptype := params["type"]
	if ptype == "" {
		ptype = "P"
	}

	name := params["name"]
	if name == "" {
		return fmt.Errorf("parametro -name requerido")
	}

	return Disk.Fdisk(size, unit, fit, path, ptype, name)
}

func processMount(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"path": true,
		"name": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para mount", param)
		}
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	name := params["name"]
	if name == "" {
		return fmt.Errorf("parametro -name requerido")
	}

	return Disk.Mount(path, name)
}

func processMkfs(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"id":     true,
		"type":   true,
		"format": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para mkfs", param)
		}
	}

	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	fsType := params["type"]
	if fsType == "" {
		fsType = "ext2"
	}

	formatType := params["format"]
	if formatType == "" {
		formatType = "full"
	}

	return Disk.Mkfs(id, fsType, formatType)
}

func parseParameters(args []string) map[string]string {
	params := make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				key := strings.TrimPrefix(parts[0], "-")
				value := strings.Trim(parts[1], "\"")
				params[key] = value
			} else {
				key := strings.TrimPrefix(arg, "-")
				params[key] = "true"
			}
		}
	}

	return params
}
