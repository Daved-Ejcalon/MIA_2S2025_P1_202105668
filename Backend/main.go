package main

import (
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== PRUEBA DEL SISTEMA DE ARCHIVOS EXT2 ===")

	if err := ejecutarPruebas(); err != nil {
		log.Fatal("Error en las pruebas:", err)
	}

	fmt.Println("\n=== TODAS LAS PRUEBAS COMPLETADAS ===")
	fmt.Println("Archivo generado: ./TestDisco.mia")

	if preguntarEliminarArchivo() {
		if err := Disk.RmDisk("./TestDisco.mia"); err != nil {
			fmt.Println("Error eliminando:", err)
		} else {
			fmt.Println("✅ Archivo de prueba eliminado")
		}
	}
}

func ejecutarPruebas() error {
	// Ejecutar todas las pruebas en secuencia
	if err := crearDisco(); err != nil {
		return err
	}

	if err := crearParticionPrimaria(); err != nil {
		return err
	}

	if err := crearParticionExtendida(); err != nil {
		return err
	}

	if err := crearParticionLogica(); err != nil {
		return err
	}

	if err := mostrarParticiones(); err != nil {
		return err
	}

	return nil
}

func crearDisco() error {
	fmt.Println("\n1. Creando disco de 50MB...")
	err := Disk.MkDisk(50, "M", "FF", "./TestDisco.mia")
	if err != nil {
		return fmt.Errorf("error creando disco: %w", err)
	}
	fmt.Println("✅ Disco creado exitosamente")
	return nil
}

func crearParticionPrimaria() error {
	fmt.Println("\n2. Creando partición primaria de 10MB...")
	err := Disk.Fdisk(10, "M", "WF", "./TestDisco.mia", "P", "Particion1")
	if err != nil {
		return fmt.Errorf("error creando partición primaria: %w", err)
	}
	fmt.Println("✅ Partición primaria creada exitosamente")
	return nil
}

func crearParticionExtendida() error {
	fmt.Println("\n3. Creando partición extendida de 15MB...")
	err := Disk.Fdisk(15, "M", "WF", "./TestDisco.mia", "E", "Extendida1")
	if err != nil {
		return fmt.Errorf("error creando partición extendida: %w", err)
	}
	fmt.Println("✅ Partición extendida creada exitosamente")
	return nil
}

func crearParticionLogica() error {
	fmt.Println("\n4. Creando partición lógica de 5MB...")
	err := Disk.Fdisk(5, "M", "WF", "./TestDisco.mia", "L", "Logica1")
	if err != nil {
		return fmt.Errorf("error creando partición lógica: %w", err)
	}
	fmt.Println("✅ Partición lógica creada exitosamente")
	return nil
}

func mostrarParticiones() error {
	fmt.Println("\n5. Mostrando información del disco...")
	err := Disk.ShowDisk("./TestDisco.mia")
	if err != nil {
		return fmt.Errorf("error mostrando particiones: %w", err)
	}
	fmt.Println("✅ Información mostrada exitosamente")
	return nil
}

func preguntarEliminarArchivo() bool {
	fmt.Print("\n¿Deseas eliminar el archivo de prueba? (s/n): ")
	reader := bufio.NewReader(os.Stdin)
	respuesta, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error leyendo respuesta, manteniendo archivo")
		return false
	}

	respuesta = strings.TrimSpace(strings.ToLower(respuesta))
	return respuesta == "s" || respuesta == "si" || respuesta == "y" || respuesta == "yes"
}
