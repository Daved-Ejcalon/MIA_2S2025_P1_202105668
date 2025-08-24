package Disk

import (
	"fmt"
	"strings"
)

// Mounted muestra todas las particiones montadas en memoria
func Mounted() {
	// Verificar si hay particiones montadas
	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas en el sistema")
		return
	}

	// Mostrar header
	fmt.Println("=== PARTICIONES MONTADAS ===")

	// Recopilar todos los IDs
	var mountedIDs []string
	for _, mount := range mountedPartitions {
		mountedIDs = append(mountedIDs, mount.MountID)
	}

	// Mostrar IDs en formato compacto (opción 1)
	fmt.Printf("IDs: %s\n", strings.Join(mountedIDs, ", "))

	fmt.Println("\n=== DETALLE DE MONTAJES ===")
	// Mostrar información detallada de cada montaje
	for _, mount := range mountedPartitions {
		fmt.Printf("ID: %s | Partición: %s | Disco: %s | Letra: %c | Número: %d\n",
			mount.MountID,
			mount.PartitionName,
			mount.DiskPath,
			mount.DiskLetter,
			mount.PartNumber)
	}

	// Mostrar total
	fmt.Printf("\nTotal de particiones montadas: %d\n", len(mountedPartitions))
}

// MountedSimple muestra solo los IDs de las particiones montadas (versión minimalista)
func MountedSimple() {
	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas")
		return
	}

	// Recopilar IDs
	var mountedIDs []string
	for _, mount := range mountedPartitions {
		mountedIDs = append(mountedIDs, mount.MountID)
	}

	// Mostrar solo los IDs separados por comas
	fmt.Printf("Particiones montadas: %s\n", strings.Join(mountedIDs, ", "))
}

// MountedTable muestra las particiones en formato tabla (versión detallada)
func MountedTable() {
	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas en el sistema")
		return
	}

	// Header de tabla
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        PARTICIONES MONTADAS                         ║")
	fmt.Println("╠════════╦═══════════════╦═══════════════════════════════╦═══════╦═══════╣")
	fmt.Println("║   ID   ║   PARTICIÓN   ║            DISCO              ║ LETRA ║   #   ║")
	fmt.Println("╠════════╬═══════════════╬═══════════════════════════════╬═══════╬═══════╣")

	// Filas de datos
	for _, mount := range mountedPartitions {
		// Truncar nombres largos para que quepan en la tabla
		partName := mount.PartitionName
		if len(partName) > 13 {
			partName = partName[:10] + "..."
		}

		diskPath := mount.DiskPath
		if len(diskPath) > 29 {
			diskPath = "..." + diskPath[len(diskPath)-26:]
		}

		fmt.Printf("║ %-6s ║ %-13s ║ %-29s ║   %c   ║   %d   ║\n",
			mount.MountID, partName, diskPath, mount.DiskLetter, mount.PartNumber)
	}

	// Footer
	fmt.Println("╚════════╩═══════════════╩═══════════════════════════════╩═══════╩═══════╝")
	fmt.Printf("Total: %d particiones montadas\n", len(mountedPartitions))
}
