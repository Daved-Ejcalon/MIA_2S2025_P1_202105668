package Graphviz

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GenerateMBRGraph genera el gráfico DOT para el reporte MBR
func GenerateMBRGraph(diskPath string, outputPath string) error {
	// Leer datos del MBR
	mbr, err := readMBRFromDisk(diskPath)
	if err != nil {
		return fmt.Errorf("error al leer MBR: %v", err)
	}

	// Generar contenido DOT
	dotContent := generateMBRDotContent(mbr, diskPath)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "mbr_report.dot")

	err = os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen JPG usando Graphviz
	return generateImageFromDot(dotFile, outputPath)
}

func readMBRFromDisk(diskPath string) (*Models.MBR, error) {
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

func generateMBRDotContent(mbr *Models.MBR, diskPath string) string {
	var dot strings.Builder

	dot.WriteString("digraph MBR {\n")
	dot.WriteString("    node [shape=plaintext, fontname=\"Arial\"];\n")
	dot.WriteString("    rankdir=TB;\n")
	dot.WriteString("    bgcolor=\"#121212\";\n")
	dot.WriteString("    dpi=3600;\n\n")

	// Crear tabla HTML con diseño minimalista
	dot.WriteString("    mbr_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#121212\">\n")

	// Header principal con cortes rectos
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#7c3aed\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#ffffff\" POINT-SIZE=\"24\"><B>REPORTE DE MBR</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString("            <TR><TD COLSPAN=\"2\" HEIGHT=\"10\"></TD></TR>\n")

	// Sección información MBR
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#1e1e1e\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#2a2a2a\">\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>mbr_tamano</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", mbr.MbrSize))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>mbr_fecha_creacion</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%s</FONT></TD>\n", time.Unix(mbr.MbrCreationDate, 0).Format("2006-01-02 15:04")))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>mbr_disk_signature</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", mbr.MbrSignature))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Procesar particiones del MBR
	for _, partition := range mbr.Partitions {
		if !partition.IsEmptyPartition() {
			// Espacio separador
			dot.WriteString("            <TR><TD COLSPAN=\"2\" HEIGHT=\"15\"></TD></TR>\n")

			// Determinar color según tipo de partición
			headerBg := "#4c1d95"
			headerText := "#e9d5ff"
			if partition.IsExtended() {
				headerBg = "#1e293b"
				headerText = "#bae6fd"
			}

			// Header de partición
			dot.WriteString("            <TR>\n")
			dot.WriteString(fmt.Sprintf("                <TD COLSPAN=\"2\" BGCOLOR=\"%s\" ALIGN=\"center\">\n", headerBg))
			if partition.IsPrimary() {
				dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"%s\" POINT-SIZE=\"14\"><B>PARTICIÓN PRIMARIA</B></FONT>\n", headerText))
			} else if partition.IsExtended() {
				dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"%s\" POINT-SIZE=\"14\"><B>PARTICIÓN EXTENDIDA</B></FONT>\n", headerText))
			} else {
				dot.WriteString(fmt.Sprintf("                    <FONT COLOR=\"%s\" POINT-SIZE=\"14\"><B>PARTICIÓN</B></FONT>\n", headerText))
			}
			dot.WriteString("                </TD>\n")
			dot.WriteString("            </TR>\n")

			// Datos de la partición
			dot.WriteString("            <TR>\n")
			dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#1e1e1e\" ALIGN=\"center\">\n")
			dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#2a2a2a\">\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_status</B></FONT></TD>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", partition.PartStatus))
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_type</B></FONT></TD>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%c</FONT></TD>\n", partition.PartType))
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_fit</B></FONT></TD>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%c</FONT></TD>\n", partition.PartFit))
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_start</B></FONT></TD>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", partition.PartStart))
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_size</B></FONT></TD>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", partition.PartSize))
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                        <TR>\n")
			dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_name</B></FONT></TD>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%s</FONT></TD>\n", partition.GetPartitionName()))
			dot.WriteString("                        </TR>\n")
			dot.WriteString("                    </TABLE>\n")
			dot.WriteString("                </TD>\n")
			dot.WriteString("            </TR>\n")

			// Si es partición extendida, leer particiones lógicas
			if partition.IsExtended() {
				logicalPartitions, err := readLogicalPartitions(diskPath, partition.PartStart)
				if err == nil && len(logicalPartitions) > 0 {
					for _, logical := range logicalPartitions {
						// Espacio separador
						dot.WriteString("            <TR><TD COLSPAN=\"2\" HEIGHT=\"10\"></TD></TR>\n")

						// Header de partición lógica
						dot.WriteString("            <TR>\n")
						dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#7f1d1d\" ALIGN=\"center\">\n")
						dot.WriteString("                    <FONT COLOR=\"#fecaca\" POINT-SIZE=\"12\"><B>PARTICIÓN LÓGICA</B></FONT>\n")
						dot.WriteString("                </TD>\n")
						dot.WriteString("            </TR>\n")

						// Datos de la partición lógica
						dot.WriteString("            <TR>\n")
						dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#1e1e1e\" ALIGN=\"center\">\n")
						dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#2a2a2a\">\n")
						dot.WriteString("                        <TR>\n")
						dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_status</B></FONT></TD>\n")
						dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", logical.PartMount))
						dot.WriteString("                        </TR>\n")
						dot.WriteString("                        <TR>\n")
						dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_next</B></FONT></TD>\n")
						if logical.PartNext == -1 {
							dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">-1</FONT></TD>\n")
						} else {
							dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", logical.PartNext))
						}
						dot.WriteString("                        </TR>\n")
						dot.WriteString("                        <TR>\n")
						dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_fit</B></FONT></TD>\n")
						dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%c</FONT></TD>\n", logical.PartFit))
						dot.WriteString("                        </TR>\n")
						dot.WriteString("                        <TR>\n")
						dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_start</B></FONT></TD>\n")
						dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", logical.PartStart))
						dot.WriteString("                        </TR>\n")
						dot.WriteString("                        <TR>\n")
						dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_size</B></FONT></TD>\n")
						dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%d</FONT></TD>\n", logical.PartS))
						dot.WriteString("                        </TR>\n")
						dot.WriteString("                        <TR>\n")
						dot.WriteString("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#9ca3af\"><B>part_name</B></FONT></TD>\n")
						dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#1e1e1e\" ALIGN=\"center\"><FONT COLOR=\"#ffffff\">%s</FONT></TD>\n", logical.GetLogicalPartitionName()))
						dot.WriteString("                        </TR>\n")
						dot.WriteString("                    </TABLE>\n")
						dot.WriteString("                </TD>\n")
						dot.WriteString("            </TR>\n")
					}
				}
			}
		}
	}

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}

func readLogicalPartitions(diskPath string, extendedStart int64) ([]Models.EBR, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logicalPartitions []Models.EBR
	currentPos := extendedStart

	for {
		// Posicionarse en el EBR actual
		_, err = file.Seek(currentPos, 0)
		if err != nil {
			break
		}

		var ebr Models.EBR
		err = binary.Read(file, binary.LittleEndian, &ebr)
		if err != nil {
			break
		}

		// Si el EBR está vacío, terminar
		if ebr.IsEmptyEBR() {
			break
		}

		logicalPartitions = append(logicalPartitions, ebr)

		// Si no hay siguiente EBR, terminar
		if !ebr.HasNext() {
			break
		}

		currentPos = ebr.PartNext
	}

	return logicalPartitions, nil
}

func generateImageFromDot(dotFile string, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	os.MkdirAll(outputDir, 0755)
	// Ejecutar comando dot para
	cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputPath)
	cmd.CombinedOutput()

	return nil
}
