package Graphviz

import (
	"MIA_2S2025_P1_202105668/Models"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DiskSegment representa un segmento del disco con su información
type DiskSegment struct {
	Type       string  // "MBR", "Primaria", "Extendida", "Libre"
	Name       string  // Nombre de la partición
	Start      int64   // Posición de inicio
	Size       int64   // Tamaño del segmento
	Percentage float64 // Porcentaje del disco
	IsExtended bool    // Si es partición extendida
	LogicalPartitions []LogicalSegment // Particiones lógicas internas
}

// LogicalSegment representa una partición lógica dentro de una extendida
type LogicalSegment struct {
	Type       string  // "EBR", "Lógica", "Libre"
	Name       string  // Nombre de la partición
	Start      int64   // Posición de inicio
	Size       int64   // Tamaño del segmento
	Percentage float64 // Porcentaje dentro de la partición extendida
}

// GenerateDiskGraph genera el gráfico DOT para el reporte de disco
func GenerateDiskGraph(diskPath string, outputPath string) error {
	// Leer datos del MBR
	mbr, err := readMBRFromDisk(diskPath)
	if err != nil {
		return fmt.Errorf("error al leer MBR: %v", err)
	}

	// Calcular el layout del disco
	diskLayout, err := calculateDiskLayout(diskPath, mbr)
	if err != nil {
		return fmt.Errorf("error calculando layout del disco: %v", err)
	}

	// Generar contenido DOT
	dotContent := generateDiskDotContent(diskLayout, diskPath)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "disk_report.dot")

	err = os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen PNG usando Graphviz
	return generateImageFromDot(dotFile, outputPath)
}

// calculateDiskLayout calcula la distribución del disco con porcentajes
func calculateDiskLayout(diskPath string, mbr *Models.MBR) ([]DiskSegment, error) {
	var segments []DiskSegment
	diskSize := mbr.MbrSize
	currentPos := int64(Models.GetMBRSize()) // Empezar después del MBR

	// Agregar segmento MBR
	mbrSegment := DiskSegment{
		Type:       "MBR",
		Name:       "MBR",
		Start:      0,
		Size:       int64(Models.GetMBRSize()),
		Percentage: (float64(Models.GetMBRSize()) / float64(diskSize)) * 100,
		IsExtended: false,
	}
	segments = append(segments, mbrSegment)

	// Ordenar particiones por posición de inicio
	partitions := make([]Models.Partition, 0)
	for _, partition := range mbr.Partitions {
		if !partition.IsEmptyPartition() {
			partitions = append(partitions, partition)
		}
	}

	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].PartStart < partitions[j].PartStart
	})

	// Procesar cada partición
	for _, partition := range partitions {
		// Agregar espacio libre antes de la partición si existe
		if currentPos < partition.PartStart {
			freeSize := partition.PartStart - currentPos
			freeSegment := DiskSegment{
				Type:       "Libre",
				Name:       "Libre",
				Start:      currentPos,
				Size:       freeSize,
				Percentage: (float64(freeSize) / float64(diskSize)) * 100,
				IsExtended: false,
			}
			segments = append(segments, freeSegment)
		}

		// Agregar la partición
		if partition.IsExtended() {
			// Procesar partición extendida
			extendedSegment := DiskSegment{
				Type:       "Extendida",
				Name:       partition.GetPartitionName(),
				Start:      partition.PartStart,
				Size:       partition.PartSize,
				Percentage: (float64(partition.PartSize) / float64(diskSize)) * 100,
				IsExtended: true,
			}

			// Obtener particiones lógicas
			logicalPartitions, err := readLogicalPartitions(diskPath, partition.PartStart)
			if err == nil {
				extendedSegment.LogicalPartitions = calculateLogicalLayout(logicalPartitions, partition)
			}

			segments = append(segments, extendedSegment)
		} else {
			// Partición primaria
			primarySegment := DiskSegment{
				Type:       "Primaria",
				Name:       partition.GetPartitionName(),
				Start:      partition.PartStart,
				Size:       partition.PartSize,
				Percentage: (float64(partition.PartSize) / float64(diskSize)) * 100,
				IsExtended: false,
			}
			segments = append(segments, primarySegment)
		}

		currentPos = partition.PartStart + partition.PartSize
	}

	// Agregar espacio libre al final si existe
	if currentPos < diskSize {
		freeSize := diskSize - currentPos
		freeSegment := DiskSegment{
			Type:       "Libre",
			Name:       "Libre",
			Start:      currentPos,
			Size:       freeSize,
			Percentage: (float64(freeSize) / float64(diskSize)) * 100,
			IsExtended: false,
		}
		segments = append(segments, freeSegment)
	}

	return segments, nil
}

// calculateLogicalLayout calcula el layout de particiones lógicas dentro de una extendida
func calculateLogicalLayout(logicalPartitions []Models.EBR, extendedPartition Models.Partition) []LogicalSegment {
	var logicalSegments []LogicalSegment
	extendedSize := extendedPartition.PartSize
	currentPos := extendedPartition.PartStart

	// Ordenar particiones lógicas por posición
	sort.Slice(logicalPartitions, func(i, j int) bool {
		return logicalPartitions[i].PartStart < logicalPartitions[j].PartStart
	})

	for _, logical := range logicalPartitions {
		// Agregar EBR
		ebrStart := logical.PartStart - int64(Models.GetEBRSize())
		if ebrStart > currentPos {
			// Espacio libre antes del EBR
			freeSize := ebrStart - currentPos
			freeSegment := LogicalSegment{
				Type:       "Libre",
				Name:       "Libre",
				Start:      currentPos,
				Size:       freeSize,
				Percentage: (float64(freeSize) / float64(extendedSize)) * 100,
			}
			logicalSegments = append(logicalSegments, freeSegment)
		}

		// EBR
		ebrSegment := LogicalSegment{
			Type:       "EBR",
			Name:       "EBR",
			Start:      ebrStart,
			Size:       int64(Models.GetEBRSize()),
			Percentage: (float64(Models.GetEBRSize()) / float64(extendedSize)) * 100,
		}
		logicalSegments = append(logicalSegments, ebrSegment)

		// Partición lógica
		logicalSegment := LogicalSegment{
			Type:       "Lógica",
			Name:       logical.GetLogicalPartitionName(),
			Start:      logical.PartStart,
			Size:       logical.PartS,
			Percentage: (float64(logical.PartS) / float64(extendedSize)) * 100,
		}
		logicalSegments = append(logicalSegments, logicalSegment)

		currentPos = logical.PartStart + logical.PartS
	}

	// Espacio libre al final de la partición extendida
	extendedEnd := extendedPartition.PartStart + extendedPartition.PartSize
	if currentPos < extendedEnd {
		freeSize := extendedEnd - currentPos
		freeSegment := LogicalSegment{
			Type:       "Libre",
			Name:       "Libre",
			Start:      currentPos,
			Size:       freeSize,
			Percentage: (float64(freeSize) / float64(extendedSize)) * 100,
		}
		logicalSegments = append(logicalSegments, freeSegment)
	}

	return logicalSegments
}

// generateDiskDotContent genera el contenido DOT para el gráfico del disco
func generateDiskDotContent(diskLayout []DiskSegment, diskPath string) string {
	var dot strings.Builder

	dot.WriteString("digraph DiskReport {\n")
	dot.WriteString("    node [shape=plaintext, fontname=\"Arial\"];\n")
	dot.WriteString("    rankdir=LR;\n")
	dot.WriteString("    bgcolor=\"#f0f0f0\";\n")
	dot.WriteString("    dpi=300;\n\n")

	// Crear tabla principal
	dot.WriteString("    disk_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"#ffffff\">\n")

	// Header del reporte
	dot.WriteString("            <TR>\n")
	dot.WriteString(fmt.Sprintf("                <TD COLSPAN=\"%d\" BGCOLOR=\"#2563eb\" ALIGN=\"center\">\n", len(diskLayout)))
	dot.WriteString("                    <FONT COLOR=\"#ffffff\" POINT-SIZE=\"16\"><B>REPORTE DE DISCO</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Fila de segmentos principales
	dot.WriteString("            <TR>\n")
	for _, segment := range diskLayout {
		color := getSegmentColor(segment.Type)
		dot.WriteString(fmt.Sprintf("                <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"%d\">\n", color, int(segment.Percentage*10)))

		if segment.IsExtended && len(segment.LogicalPartitions) > 0 {
			// Partición extendida con sub-tabla
			dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString(fmt.Sprintf("                            <TD COLSPAN=\"%d\" BGCOLOR=\"%s\" ALIGN=\"center\">\n", len(segment.LogicalPartitions), color))
			dot.WriteString(fmt.Sprintf("                                <FONT COLOR=\"#ffffff\" POINT-SIZE=\"10\"><B>%s</B></FONT>\n", segment.Type))
			dot.WriteString("                            </TD>\n")
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                        <TR>\n")

			// Particiones lógicas
			for _, logical := range segment.LogicalPartitions {
				logicalColor := getSegmentColor(logical.Type)
				dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"%d\">\n", logicalColor, int(logical.Percentage*2)))
				dot.WriteString(fmt.Sprintf("                                <FONT COLOR=\"#000000\" POINT-SIZE=\"8\"><B>%s</B></FONT>\n", logical.Type))
				dot.WriteString("                            </TD>\n")
			}

			dot.WriteString("                        </TR>\n")
			dot.WriteString("                    </TABLE>\n")
		} else {
			// Segmento simple
			dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"#000000\" POINT-SIZE=\"12\"><B>%s</B></FONT>\n", segment.Type))
		}

		dot.WriteString("                </TD>\n")
	}
	dot.WriteString("            </TR>\n")

	// Fila de porcentajes
	dot.WriteString("            <TR>\n")
	for _, segment := range diskLayout {
		dot.WriteString("                <TD BGCOLOR=\"#e5e7eb\" ALIGN=\"center\">\n")
		dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"#000000\" POINT-SIZE=\"10\">%.1f%%</FONT>\n", segment.Percentage))
		dot.WriteString("                </TD>\n")
	}
	dot.WriteString("            </TR>\n")

	// Fila de nombres
	dot.WriteString("            <TR>\n")
	for _, segment := range diskLayout {
		dot.WriteString("                <TD BGCOLOR=\"#f3f4f6\" ALIGN=\"center\">\n")
		if segment.Name != "" && segment.Name != segment.Type {
			dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"#000000\" POINT-SIZE=\"9\">%s</FONT>\n", segment.Name))
		} else {
			dot.WriteString("                    <FONT COLOR=\"#000000\" POINT-SIZE=\"9\">-</FONT>\n")
		}
		dot.WriteString("                </TD>\n")
	}
	dot.WriteString("            </TR>\n")

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}

// getSegmentColor retorna el color apropiado para cada tipo de segmento
func getSegmentColor(segmentType string) string {
	switch segmentType {
	case "MBR":
		return "#1f2937"      // Gris oscuro
	case "Primaria":
		return "#3b82f6"      // Azul
	case "Extendida":
		return "#10b981"      // Verde
	case "Lógica":
		return "#f59e0b"      // Amarillo/naranja
	case "EBR":
		return "#ef4444"      // Rojo
	case "Libre":
		return "#d1d5db"      // Gris claro
	default:
		return "#9ca3af"      // Gris medio
	}
}

