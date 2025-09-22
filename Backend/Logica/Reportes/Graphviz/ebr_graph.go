package Graphviz

import (
	"MIA_2S2025_P1_202105668/Models"
	"MIA_2S2025_P1_202105668/Utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateEBRGraph genera el gráfico DOT para el reporte EBR
func GenerateEBRGraph(diskPath string, ebrName string, outputPath string) error {
	// Buscar el EBR específico
	ebr, err := findEBRByName(diskPath, ebrName)
	if err != nil {
		return fmt.Errorf("error buscando EBR '%s': %v", ebrName, err)
	}

	// Generar contenido DOT
	dotContent := generateEBRDotContent(ebr, ebrName)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "ebr_report.dot")

	err = os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen JPG usando Graphviz
	return Utils.GenerateImageFromDot(dotFile, outputPath)
}

// findEBRByName busca un EBR específico por nombre en todas las particiones extendidas
func findEBRByName(diskPath string, ebrName string) (*Models.EBR, error) {
	// Primero leer el MBR para encontrar particiones extendidas
	mbr, err := ReadMBRFromDisk(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Buscar en cada partición extendida
	for _, partition := range mbr.Partitions {
		if partition.IsExtended() {
			logicalPartitions, err := Utils.ReadLogicalPartitions(diskPath, partition.PartStart)
			if err != nil {
				continue
			}

			// Buscar por nombre en las particiones lógicas
			for _, logical := range logicalPartitions {
				if logical.GetLogicalPartitionName() == ebrName {
					return &logical, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("EBR con nombre '%s' no encontrado", ebrName)
}


// generateEBRDotContent genera el contenido DOT específico para EBR
func generateEBRDotContent(ebr *Models.EBR, ebrName string) string {
	var dot strings.Builder

	// Usar configuración base con altura fija para EBR (más pequeño que MBR)
	totalHeight := 6.0
	dot.WriteString(Utils.GetBaseGraphConfig(totalHeight))

	// Crear tabla HTML con diseño minimalista
	dot.WriteString("    ebr_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#2a2a2a\">\n")

	// Header principal
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#991b1b\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"24\"><B>REPORTE DE EBR</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("10"))

	// Título de la partición lógica
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#7f1d1d\" ALIGN=\"center\">\n")
	dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"#fecaca\" POINT-SIZE=\"16\"><B>PARTICIÓN LÓGICA: %s</B></FONT>\n", ebrName))
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Datos del EBR
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString(Utils.GetTableWrapperStart())

	// part_mount (equivalente a part_status en EBR)
	dot.WriteString(Utils.GetTableRowStyle("part_mount", fmt.Sprintf("%d", ebr.PartMount)))

	// part_fit
	dot.WriteString(Utils.GetTableRowStyle("part_fit", fmt.Sprintf("%c", ebr.PartFit)))

	// part_start
	dot.WriteString(Utils.GetTableRowStyle("part_start", fmt.Sprintf("%d", ebr.PartStart)))

	// part_size
	dot.WriteString(Utils.GetTableRowStyle("part_size", fmt.Sprintf("%d", ebr.PartS)))

	// part_next
	nextValue := "-1"
	if ebr.HasNext() {
		nextValue = fmt.Sprintf("%d", ebr.PartNext)
	}
	dot.WriteString(Utils.GetTableRowStyle("part_next", nextValue))

	// part_name
	dot.WriteString(Utils.GetTableRowStyle("part_name", ebr.GetLogicalPartitionName()))

	dot.WriteString(Utils.GetTableWrapperEnd())
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}

// GenerateEBRCompleteGraph genera el gráfico DOT para un reporte completo de todos los EBRs
func GenerateEBRCompleteGraph(diskPath string, outputPath string) error {
	// Leer el MBR para encontrar particiones extendidas
	mbr, err := ReadMBRFromDisk(diskPath)
	if err != nil {
		return fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Recopilar todas las particiones lógicas
	var allLogicalPartitions []Models.EBR
	var extendedPartitionNames []string

	for _, partition := range mbr.Partitions {
		if partition.IsExtended() {
			logicalPartitions, err := Utils.ReadLogicalPartitions(diskPath, partition.PartStart)
			if err == nil && len(logicalPartitions) > 0 {
				allLogicalPartitions = append(allLogicalPartitions, logicalPartitions...)
				extendedPartitionNames = append(extendedPartitionNames, partition.GetPartitionName())
			}
		}
	}

	if len(allLogicalPartitions) == 0 {
		return fmt.Errorf("no se encontraron particiones lógicas en el disco")
	}

	// Generar contenido DOT
	dotContent := generateEBRCompleteDotContent(allLogicalPartitions, extendedPartitionNames)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "ebr_complete_report.dot")

	err = os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen JPG usando Graphviz
	return Utils.GenerateImageFromDot(dotFile, outputPath)
}

// generateEBRCompleteDotContent genera el contenido DOT para reporte completo de EBRs
func generateEBRCompleteDotContent(logicalPartitions []Models.EBR, extendedNames []string) string {
	var dot strings.Builder

	// Calcular altura dinámica basada en el número de particiones lógicas
	totalHeight := Utils.CalculateGraphHeight(0, len(logicalPartitions)) + 2.0
	dot.WriteString(Utils.GetBaseGraphConfig(totalHeight))

	// Crear tabla HTML con diseño minimalista
	dot.WriteString("    ebr_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#2a2a2a\">\n")

	// Header principal
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#991b1b\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"24\"><B>REPORTE COMPLETO DE EBRs</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("15"))

	// Información general
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString(Utils.GetTableWrapperStart())
	dot.WriteString(Utils.GetTableRowStyle("Total Particiones Lógicas", fmt.Sprintf("%d", len(logicalPartitions))))
	dot.WriteString(Utils.GetTableRowStyle("Particiones Extendidas", fmt.Sprintf("%d", len(extendedNames))))
	dot.WriteString(Utils.GetTableWrapperEnd())
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Mostrar cada partición lógica
	for i, logical := range logicalPartitions {
		// Espacio separador
		dot.WriteString(Utils.GetSeparatorRow("15"))

		// Header de partición lógica
		dot.WriteString("            <TR>\n")
		dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#7f1d1d\" ALIGN=\"center\">\n")
		dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"#fecaca\" POINT-SIZE=\"16\"><B>PARTICIÓN LÓGICA %d: %s</B></FONT>\n", i+1, logical.GetLogicalPartitionName()))
		dot.WriteString("                </TD>\n")
		dot.WriteString("            </TR>\n")

		// Datos del EBR
		dot.WriteString("            <TR>\n")
		dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
		dot.WriteString(Utils.GetTableWrapperStart())

		// part_mount
		dot.WriteString(Utils.GetTableRowStyle("part_mount", fmt.Sprintf("%d", logical.PartMount)))

		// part_fit
		dot.WriteString(Utils.GetTableRowStyle("part_fit", fmt.Sprintf("%c", logical.PartFit)))

		// part_start
		dot.WriteString(Utils.GetTableRowStyle("part_start", fmt.Sprintf("%d", logical.PartStart)))

		// part_size
		dot.WriteString(Utils.GetTableRowStyle("part_size", fmt.Sprintf("%d", logical.PartS)))

		// part_next
		nextValue := "-1"
		if logical.HasNext() {
			nextValue = fmt.Sprintf("%d", logical.PartNext)
		}
		dot.WriteString(Utils.GetTableRowStyle("part_next", nextValue))

		// part_name
		dot.WriteString(Utils.GetTableRowStyle("part_name", logical.GetLogicalPartitionName()))

		dot.WriteString(Utils.GetTableWrapperEnd())
		dot.WriteString("                </TD>\n")
		dot.WriteString("            </TR>\n")
	}

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}
