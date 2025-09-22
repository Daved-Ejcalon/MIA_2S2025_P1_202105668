package main

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Reportes"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Logica/Users/Comandos"
	"MIA_2S2025_P1_202105668/Logica/Users/Root"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Verificar si se debe ejecutar como servidor web
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
		return
	}

	// Modo consola tradicional
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

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("PANIC: %v\n", r)
				}
			}()
			err := processCommand(input)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}
		}()
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
		return processCat(params)
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
	case "rep":
		return processRep(params)
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
		unit = "M"
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
		"id":   true,
		"type": true,
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

	return Disk.Mkfs(id)
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

func processCat(params map[string]string) error {
	// Verificar sesión activa
	session := Users.GetCurrentSession()
	if session == nil || !session.IsActive {
		return fmt.Errorf("ERROR: Debe iniciar sesión para usar este comando")
	}

	// Pasar la sesión al comando Cat
	return Disk.CatWithSession(params, session.MountID)
}

func processRep(params map[string]string) error {
	// Validar que solo se usen parámetros permitidos
	validParams := map[string]bool{
		"name":         true,
		"path":         true,
		"id":           true,
		"path_file_ls": true,
	}

	for param := range params {
		if !validParams[param] {
			return fmt.Errorf("parametro -%s no es valido para rep", param)
		}
	}

	// Validar parámetros obligatorios
	name := params["name"]
	if name == "" {
		return fmt.Errorf("parametro -name requerido")
	}

	path := params["path"]
	if path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	id := params["id"]
	if id == "" {
		return fmt.Errorf("parametro -id requerido")
	}

	// Validar valores válidos para name
	validNames := map[string]bool{
		"mbr":   true,
		"disk":  true,
		"ebr":   true,
		"inode": true,
		"sb":    true,
		"file":  true,
		"ls":    true,
	}

	if !validNames[name] {
		return fmt.Errorf("valor de -name debe ser: mbr, disk, ebr, inode, sb, file o ls")
	}

	// Llamar al generador de reportes correspondiente
	return Reportes.GenerateReport(name, id, path, params["path_file_ls"])
}

// === SERVIDOR WEB ===

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	// Capturar la salida estándar
	oldStdout := os.Stdout
	r_out, w_out, _ := os.Pipe()
	os.Stdout = w_out

	// Capturar errores
	var cmdError error

	// Ejecutar el comando en una goroutine para capturar la salida
	done := make(chan bool)
	var output []byte

	go func() {
		defer func() {
			if r := recover(); r != nil {
				cmdError = fmt.Errorf("PANIC: %v", r)
			}
			done <- true
		}()

		// Procesar el comando usando la función existente
		cmdError = processCommand(strings.TrimSpace(req.Command))
	}()

	// Esperar a que termine y cerrar el pipe
	go func() {
		<-done
		w_out.Close()
	}()

	// Leer la salida
	output, _ = io.ReadAll(r_out)
	os.Stdout = oldStdout

	// Preparar la respuesta
	resp := CommandResponse{
		Output: string(output),
	}

	if cmdError != nil {
		resp.Error = cmdError.Error()
		w.WriteHeader(http.StatusBadRequest)
	}

	// Enviar respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func startServer() {
	http.HandleFunc("/execute", executeCommandHandler)

	fmt.Println("Ctrl+C")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
