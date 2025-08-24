package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

// ShowDisk muestra información detallada del disco y sus particiones
func ShowDisk(path string) error {
	// Verificar que el archivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("el archivo %s no existe", path)
	}

	// Abrir archivo del disco
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco: %w", err)
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return fmt.Errorf("error leyendo MBR: %w", err)
	}

	// Mostrar información del MBR
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("               INFORMACIÓN DEL DISCO")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("📁 Archivo: %s\n", path)
	fmt.Printf("💾 Tamaño total: %d bytes (%.2f MB)\n",
		mbr.MbrSize, float64(mbr.MbrSize)/(1024*1024))
	fmt.Printf("📅 Fecha de creación: %s\n",
		time.Unix(mbr.MbrCreationDate, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("🔑 Signature: %d\n", mbr.MbrSignature)
	fmt.Printf("⚙️  Algoritmo de ajuste: %c\n", mbr.DiskFit)

	// Mostrar información de particiones
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("                  PARTICIONES")
	fmt.Println(strings.Repeat("-", 60))

	partitionCount := 0
	usedSpace := int64(0)

	for i, partition := range mbr.Partitions {
		if partition.PartStatus != 0 {
			partitionCount++
			usedSpace += partition.PartSize

			fmt.Printf("\n🔸 Partición %d:\n", i+1)
			fmt.Printf("   Nombre: %s\n", partition.GetName())
			fmt.Printf("   Tipo: %s\n", getPartitionTypeString(partition.PartType))
			fmt.Printf("   Estado: %s\n", getPartitionStatusString(partition.PartStatus))
			fmt.Printf("   Ajuste: %s\n", getFitString(partition.PartFit))
			fmt.Printf("   Inicio: %d bytes\n", partition.PartStart)
			fmt.Printf("   Tamaño: %d bytes (%.2f MB)\n",
				partition.PartSize, float64(partition.PartSize)/(1024*1024))
			fmt.Printf("   Correlativo: %d\n", partition.PartCorrelative)

			// Mostrar porcentaje del disco
			percentage := float64(partition.PartSize) / float64(mbr.MbrSize) * 100
			fmt.Printf("   Porcentaje del disco: %.2f%%\n", percentage)
		}
	}

	// Mostrar estadísticas generales
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("                  ESTADÍSTICAS")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Printf("📊 Total de particiones: %d/4\n", partitionCount)
	fmt.Printf("💿 Espacio utilizado: %d bytes (%.2f MB)\n",
		usedSpace, float64(usedSpace)/(1024*1024))

	freeSpace := mbr.MbrSize - usedSpace
	fmt.Printf("🆓 Espacio libre: %d bytes (%.2f MB)\n",
		freeSpace, float64(freeSpace)/(1024*1024))

	usedPercentage := float64(usedSpace) / float64(mbr.MbrSize) * 100
	fmt.Printf("📈 Porcentaje utilizado: %.2f%%\n", usedPercentage)

	fmt.Println(strings.Repeat("=", 60))

	return nil
}

// Funciones auxiliares para mostrar información legible
func getPartitionTypeString(partType byte) string {
	switch partType {
	case 'P':
		return "Primaria"
	case 'E':
		return "Extendida"
	case 'L':
		return "Lógica"
	default:
		return "Desconocido"
	}
}

func getPartitionStatusString(status byte) string {
	switch status {
	case 0:
		return "Inactiva"
	case 1:
		return "Activa"
	default:
		return "Desconocido"
	}
}

func getFitString(fit byte) string {
	switch fit {
	case 'B':
		return "Best Fit"
	case 'F':
		return "First Fit"
	case 'W':
		return "Worst Fit"
	default:
		return "Desconocido"
	}
}
