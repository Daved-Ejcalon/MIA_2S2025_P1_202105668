package Graphviz

import (
	"MIA_2S2025_P1_202105668/Models"
	"MIA_2S2025_P1_202105668/Utils"
	"encoding/binary"
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
	mbr, _ := ReadMBRFromDisk(diskPath)
	diskLayout, _ := calculateDiskLayout(diskPath, mbr)
	dotContent := generateDiskDotContent(diskLayout, diskPath)

	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "disk_report.dot")
	os.WriteFile(dotFile, []byte(dotContent), 0644)
	defer os.Remove(dotFile)

	return Utils.GenerateImageFromDot(dotFile, outputPath)
}

// ReadMBRFromDisk lee el MBR desde el disco (función específica para disk_graph)
func ReadMBRFromDisk(diskPath string) (*Models.MBR, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	mbr := &Models.MBR{}
	err = binary.Read(file, binary.LittleEndian, mbr)
	if err != nil {
		return nil, err
	}

	return mbr, nil
}


// calculateDiskLayout calcula la distribución del disco con porcentajes
func calculateDiskLayout(diskPath string, mbr *Models.MBR) ([]DiskSegment, error) {
	var segments []DiskSegment
	diskSize := mbr.MbrSize
	currentPos := int64(Models.GetMBRSize())


	mbrSegment := DiskSegment{
		Type:       "MBR",
		Name:       "MBR",
		Start:      0,
		Size:       int64(Models.GetMBRSize()),
		Percentage: (float64(Models.GetMBRSize()) / float64(diskSize)) * 100,
		IsExtended: false,
	}
	segments = append(segments, mbrSegment)

	partitions := make([]Models.Partition, 0)
	for _, partition := range mbr.Partitions {
		if !partition.IsEmptyPartition() {
			partitions = append(partitions, partition)
		}
	}

	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].PartStart < partitions[j].PartStart
	})

	for _, partition := range partitions {
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

		if partition.IsExtended() {
			extendedSegment := DiskSegment{
				Type:       "Extendida",
				Name:       partition.GetPartitionName(),
				Start:      partition.PartStart,
				Size:       partition.PartSize,
				Percentage: (float64(partition.PartSize) / float64(diskSize)) * 100,
				IsExtended: true,
			}


			logicalPartitions, _ := Utils.ReadLogicalPartitions(diskPath, partition.PartStart)
			extendedSegment.LogicalPartitions = calculateLogicalLayout(logicalPartitions, partition)
			segments = append(segments, extendedSegment)
		} else {
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

	sort.Slice(logicalPartitions, func(i, j int) bool {
		return logicalPartitions[i].PartStart < logicalPartitions[j].PartStart
	})

	for _, logical := range logicalPartitions {
		// Validar que tenga datos válidos (PartStart > 0 indica partición válida)
		if logical.PartStart <= 0 || logical.PartS <= 0 {
			continue
		}

		ebrStart := logical.PartStart - int64(Models.GetEBRSize())

		// Espacio libre antes del EBR si existe
		if ebrStart > currentPos {
			freeSize := ebrStart - currentPos
			if freeSize > 0 {
				freeSegment := LogicalSegment{
					Type:       "Libre",
					Name:       "Libre",
					Start:      currentPos,
					Size:       freeSize,
					Percentage: (float64(freeSize) / float64(extendedSize)) * 100,
				}
				logicalSegments = append(logicalSegments, freeSegment)
			}
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

	// Espacio libre al final
	extendedEnd := extendedPartition.PartStart + extendedPartition.PartSize
	if currentPos < extendedEnd {
		freeSize := extendedEnd - currentPos
		if freeSize > 0 {
			freeSegment := LogicalSegment{
				Type:       "Libre",
				Name:       "Libre",
				Start:      currentPos,
				Size:       freeSize,
				Percentage: (float64(freeSize) / float64(extendedSize)) * 100,
			}
			logicalSegments = append(logicalSegments, freeSegment)
		}
	}

	return logicalSegments
}

// generateDiskDotContent genera el contenido DOT para el gráfico del disco
func generateDiskDotContent(diskLayout []DiskSegment, diskPath string) string {
	var dot strings.Builder

	dot.WriteString("digraph ReporteDisco {\n")
	dot.WriteString("    node [shape=plaintext, fontname=\"Arial\"];\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    bgcolor=\"#2a2a2a\";\n")
	dot.WriteString("    dpi=1500;\n")
	dot.WriteString("    margin=0;\n\n")

	// Calcular información del disco
	diskName := filepath.Base(diskPath)
	totalSize := int64(0)
	for _, segment := range diskLayout {
		totalSize += segment.Size
	}

	// Verificar si hay particiones extendidas con lógicas
	hasExtendedWithLogicals := false
	for _, segment := range diskLayout {
		if segment.IsExtended && len(segment.LogicalPartitions) > 0 {
			hasExtendedWithLogicals = true
			break
		}
	}

	dot.WriteString("    disk [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#2a2a2a\">\n")

	// Header principal con información del disco (similar al MBR)
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"6\" BGCOLOR=\"#5b21b6\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"24\"><B>REPORTE DE DISCO</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Información del disco
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"6\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#4a4a4a\">\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Archivo</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", diskName))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Tamaño Total</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%.1f MB</FONT></TD>\n", float64(totalSize)/(1024*1024)))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString("            <TR><TD COLSPAN=\"6\" HEIGHT=\"15\"></TD></TR>\n")

	// Comenzar tabla de particiones
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"6\">\n")
	dot.WriteString("                    <table border=\"2\" cellborder=\"1\" cellspacing=\"0\">\n")

	if hasExtendedWithLogicals {
		// Formato complejo para particiones extendidas con lógicas

		// Fila 1: Tipos de partición principales
		dot.WriteString("            <tr>\n")
		for _, segment := range diskLayout {
			if segment.Type == "MBR" {
				continue
			}

			color := getSegmentColor(segment.Type)
			sizeInMB := float64(segment.Size) / (1024 * 1024)

			if segment.IsExtended && len(segment.LogicalPartitions) > 0 {
				// Calcular colspan dinámicamente basado en número de elementos lógicos
				logicalCells := 0
				for _, logical := range segment.LogicalPartitions {
					if logical.Type == "EBR" || logical.Type == "Lógica" || (logical.Type == "Libre" && logical.Size > 0) {
						logicalCells++
					}
				}

				dot.WriteString(fmt.Sprintf("                <td colspan=\"%d\" bgcolor=\"%s\"><font color=\"white\"><b>Extendida - %s (%.0fMB - %.0f%%)</b></font></td>\n",
					logicalCells, color, segment.Name, sizeInMB, segment.Percentage))
			} else {
				typeName := "Primaria"
				if segment.Type == "Libre" {
					typeName = "Libre"
				}
				dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\"><font color=\"white\"><b>%s</b></font></td>\n", color, typeName))
			}
		}
		dot.WriteString("            </tr>\n")

		// Fila 2: Detalles de particiones lógicas y información de primarias/libre
		dot.WriteString("            <tr>\n")
		for _, segment := range diskLayout {
			if segment.Type == "MBR" {
				continue
			}

			if segment.IsExtended && len(segment.LogicalPartitions) > 0 {
				// Mostrar detalles de particiones lógicas
				for _, logical := range segment.LogicalPartitions {
					logicalColor := getSegmentColor(logical.Type)
					logicalSizeInMB := float64(logical.Size) / (1024 * 1024)

					if logical.Type == "EBR" {
						dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\" width=\"30\"><font color=\"black\"><b>EBR</b></font></td>\n", logicalColor))
					} else if logical.Type == "Lógica" {
						dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\" width=\"60\"><font color=\"black\"><b>%s<br/>%.0fMB</b></font></td>\n",
							logicalColor, logical.Name, logicalSizeInMB))
					} else if logical.Type == "Libre" && logical.Size > 0 {
						dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\" width=\"50\"><font color=\"black\"><b>Libre<br/>%.0fMB</b></font></td>\n",
							logicalColor, logicalSizeInMB))
					}
				}
			} else {
				// Partición simple (primaria o libre)
				color := getSegmentColor(segment.Type)
				sizeInMB := float64(segment.Size) / (1024 * 1024)

				if segment.Type == "Libre" {
					dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\"><font color=\"white\"><b>%.0fMB<br/>%.0f%%</b></font></td>\n",
						color, sizeInMB, segment.Percentage))
				} else {
					dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\"><font color=\"white\"><b>%s<br/>%.0fMB</b></font></td>\n",
						color, segment.Name, sizeInMB))
				}
			}
		}
		dot.WriteString("            </tr>\n")

		// Fila 3: Información descriptiva
		dot.WriteString("            <tr>\n")
		for _, segment := range diskLayout {
			if segment.Type == "MBR" {
				continue
			}

			if segment.IsExtended && len(segment.LogicalPartitions) > 0 {
				// Contar particiones lógicas
				logicalCount := 0
				for _, logical := range segment.LogicalPartitions {
					if logical.Type == "Lógica" {
						logicalCount++
					}
				}

				for _, logical := range segment.LogicalPartitions {
					if logical.Type == "EBR" {
						dot.WriteString("                <td bgcolor=\"#3a3a3a\"><font color=\"#e5e7eb\" point-size=\"9\">EBR</font></td>\n")
					} else if logical.Type == "Lógica" {
						dot.WriteString(fmt.Sprintf("                <td bgcolor=\"#4a4a4a\"><font color=\"#e5e7eb\" point-size=\"9\">%s</font></td>\n", logical.Name))
					} else if logical.Type == "Libre" && logical.Size > 0 {
						dot.WriteString("                <td bgcolor=\"#1e1e1e\"><font color=\"#9ca3af\" point-size=\"9\">Libre</font></td>\n")
					}
				}
			} else {
				if segment.Type == "Libre" {
					dot.WriteString("                <td bgcolor=\"#1e1e1e\"><font color=\"#9ca3af\" point-size=\"10\">Espacio disponible</font></td>\n")
				} else {
					dot.WriteString(fmt.Sprintf("                <td bgcolor=\"#4a4a4a\"><font color=\"#e5e7eb\" point-size=\"10\">%s</font></td>\n", segment.Name))
				}
			}
		}
		dot.WriteString("            </tr>\n")

	} else {
		// Formato simple para solo particiones primarias (como Disco 1)

		// Fila 1: Nombres
		dot.WriteString("            <tr>\n")
		for _, segment := range diskLayout {
			if segment.Type == "MBR" {
				continue
			}

			color := getSegmentColor(segment.Type)
			if segment.Type == "Libre" {
				dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\"><font color=\"white\"><b>Libre</b></font></td>\n", color))
			} else {
				dot.WriteString(fmt.Sprintf("                <td bgcolor=\"%s\"><font color=\"white\"><b>%s</b></font></td>\n", color, segment.Name))
			}
		}
		dot.WriteString("            </tr>\n")

		// Fila 2: Tamaños
		dot.WriteString("            <tr>\n")
		for _, segment := range diskLayout {
			if segment.Type == "MBR" {
				continue
			}

			sizeInMB := float64(segment.Size) / (1024 * 1024)
			dot.WriteString(fmt.Sprintf("                <td bgcolor=\"#3a3a3a\" align=\"center\"><font color=\"#e5e7eb\"><b>%.0fMB</b></font></td>\n", sizeInMB))
		}
		dot.WriteString("            </tr>\n")

		// Fila 3: Porcentajes
		dot.WriteString("            <tr>\n")
		for _, segment := range diskLayout {
			if segment.Type == "MBR" {
				continue
			}

			dot.WriteString(fmt.Sprintf("                <td bgcolor=\"#4a4a4a\" align=\"center\"><font color=\"#e5e7eb\"><b>%.1f%%</b></font></td>\n", segment.Percentage))
		}
		dot.WriteString("            </tr>\n")
	}

	dot.WriteString("                    </table>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")
	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}

// getSegmentColor retorna el color apropiado para cada tipo de segmento (coincide con MBR)
func getSegmentColor(segmentType string) string {
	switch segmentType {
	case "MBR":
		return "#6b7280"      // Gris oscuro
	case "Primaria":
		return "#4c1d95"      // Morado oscuro (igual que MBR)
	case "Extendida":
		return "#1e293b"      // Azul oscuro (igual que MBR)
	case "Lógica":
		return "#7f1d1d"      // Rojo oscuro (igual que MBR)
	case "EBR":
		return "#374151"      // Gris medio para EBRs
	case "Libre":
		return "#6b7280"      // Gris para espacio libre
	default:
		return "#333333"      // Gris medio
	}
}