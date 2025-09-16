package Disk

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

func ShowDisk(diskArgs map[string]string) error {
	path, exists := diskArgs["path"]
	if !exists || path == "" {
		return fmt.Errorf("parametro -path requerido")
	}

	fmt.Printf("=== ANÁLISIS DE DISCO ===\n")
	fmt.Printf("Ruta: %s\n", path)

	// Verificar si el archivo existe
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Printf("❌ ESTADO: Archivo no existe\n")
		return fmt.Errorf("disco no encontrado")
	}
	if err != nil {
		fmt.Printf("❌ ESTADO: Error accediendo archivo (%v)\n", err)
		return err
	}

	fmt.Printf("✅ ESTADO: Disco encontrado\n")
	fmt.Printf("📊 TAMAÑO DE ARCHIVO: %d bytes (%.2f MB)\n", fileInfo.Size(), float64(fileInfo.Size())/(1024*1024))
	fmt.Printf("📅 FECHA MODIFICACIÓN: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

	// Abrir y leer MBR
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("❌ ERROR: No se pudo abrir el disco (%v)\n", err)
		return err
	}
	defer file.Close()

	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Printf("❌ ERROR: No se pudo leer MBR (%v)\n", err)
		return err
	}

	fmt.Printf("\n=== INFORMACIÓN DEL MBR ===\n")
	fmt.Printf("💽 TAMAÑO DEL DISCO: %d bytes (%.2f MB)\n", mbr.MbrSize, float64(mbr.MbrSize)/(1024*1024))
	fmt.Printf("📅 FECHA CREACIÓN: %s\n", time.Unix(mbr.MbrCreationDate, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("🔢 SIGNATURE: %d\n", mbr.MbrSignature)
	fmt.Printf("⚙️  ALGORITMO FIT: %c\n", mbr.DiskFit)
	fmt.Printf("📏 TAMAÑO MBR: %d bytes\n", Models.GetMBRSize())

	// Verificar consistencia de tamaño
	if mbr.MbrSize != fileInfo.Size() {
		fmt.Printf("⚠️  ADVERTENCIA: Tamaño en MBR (%d) no coincide con archivo (%d)\n", mbr.MbrSize, fileInfo.Size())
	}

	// Analizar particiones
	fmt.Printf("\n=== TABLA DE PARTICIONES ===\n")
	primaryCount := 0
	extendedCount := 0
	mountedCount := 0
	var totalPartitionSize int64 = 0

	for i, partition := range mbr.Partitions {
		fmt.Printf("\n--- PARTICIÓN %d ---\n", i+1)
		
		if partition.PartStatus == 0 && partition.PartStart == 0 && partition.PartSize == 0 {
			fmt.Printf("📍 ESTADO: Vacía\n")
			continue
		}

		// Estado de la partición
		if partition.PartStatus == 1 {
			fmt.Printf("✅ ESTADO: Activa/Montada\n")
			mountedCount++
		} else {
			fmt.Printf("💤 ESTADO: Inactiva\n")
		}

		// Tipo de partición
		switch partition.PartType {
		case 'P':
			fmt.Printf("🔵 TIPO: Primaria\n")
			primaryCount++
		case 'E':
			fmt.Printf("🟡 TIPO: Extendida\n")
			extendedCount++
		case 'L':
			fmt.Printf("🟢 TIPO: Lógica\n")
		default:
			fmt.Printf("❓ TIPO: Desconocido (%c)\n", partition.PartType)
		}

		// Información de la partición
		fmt.Printf("📛 NOMBRE: %s\n", partition.GetPartitionName())
		fmt.Printf("📏 TAMAÑO: %d bytes (%.2f MB)\n", partition.PartSize, float64(partition.PartSize)/(1024*1024))
		fmt.Printf("📍 INICIO: byte %d\n", partition.PartStart)
		fmt.Printf("🔚 FIN: byte %d\n", partition.PartStart+partition.PartSize-1)
		fmt.Printf("⚙️  FIT: %c\n", partition.PartFit)

		// Información de montaje
		if partition.PartCorrelative != -1 {
			fmt.Printf("🆔 ID MONTAJE: %s\n", partition.GetPartitionID())
			fmt.Printf("🔢 CORRELATIVO: %d\n", partition.PartCorrelative)
		}

		totalPartitionSize += partition.PartSize
	}

	// Resumen
	fmt.Printf("\n=== RESUMEN ===\n")
	totalPartitions := primaryCount + extendedCount
	fmt.Printf("📊 TOTAL PARTICIONES: %d\n", totalPartitions)
	fmt.Printf("   🔵 Primarias: %d\n", primaryCount)
	fmt.Printf("   🟡 Extendidas: %d\n", extendedCount)
	fmt.Printf("   🟢 Montadas: %d\n", mountedCount)
	
	// Espacio utilizado vs disponible
	usedSpace := int64(Models.GetMBRSize()) + totalPartitionSize
	availableSpace := mbr.MbrSize - usedSpace
	
	fmt.Printf("💾 ESPACIO TOTAL: %.2f MB\n", float64(mbr.MbrSize)/(1024*1024))
	fmt.Printf("📦 ESPACIO USADO: %.2f MB (%.1f%%)\n", 
		float64(usedSpace)/(1024*1024), 
		float64(usedSpace)/float64(mbr.MbrSize)*100)
	fmt.Printf("🆓 ESPACIO LIBRE: %.2f MB (%.1f%%)\n", 
		float64(availableSpace)/(1024*1024), 
		float64(availableSpace)/float64(mbr.MbrSize)*100)

	// Validaciones
	fmt.Printf("\n=== VALIDACIONES ===\n")
	if primaryCount > 4 {
		fmt.Printf("❌ ERROR: Más de 4 particiones primarias (%d)\n", primaryCount)
	}
	if extendedCount > 1 {
		fmt.Printf("❌ ERROR: Más de 1 partición extendida (%d)\n", extendedCount)
	}
	if primaryCount + extendedCount > 4 {
		fmt.Printf("❌ ERROR: Límite de particiones MBR excedido (%d/4)\n", primaryCount + extendedCount)
	}
	if availableSpace < 0 {
		fmt.Printf("❌ ERROR: Particiones sobrepasan el tamaño del disco\n")
	}
	
	if primaryCount <= 4 && extendedCount <= 1 && primaryCount + extendedCount <= 4 && availableSpace >= 0 {
		fmt.Printf("✅ ESTADO: Disco válido\n")
	}

	return nil
}
