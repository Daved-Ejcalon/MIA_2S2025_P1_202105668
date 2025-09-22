package Utils

import (
	"MIA_2S2025_P1_202105668/Models"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateImageFromDot genera una imagen desde un archivo DOT usando Graphviz
func GenerateImageFromDot(dotFile string, outputPath string) error {
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	// Determinar formato basado en extensión
	format := "jpg"
	if strings.HasSuffix(strings.ToLower(outputPath), ".png") {
		format = "png"
	}

	cmd := exec.Command("dot", "-T"+format, "-Gdpi=1500", "-Gmargin=0", "-Gpad=0", dotFile, "-o", outputPath)
	cmd.Run()

	return nil
}

// ReadLogicalPartitions lee todas las particiones lógicas desde una partición extendida
func ReadLogicalPartitions(diskPath string, extendedStart int64) ([]Models.EBR, error) {
	file, _ := os.Open(diskPath)
	defer file.Close()

	var logicalPartitions []Models.EBR


	// Seguir la cadena de EBRs correctamente
	currentEBRPos := extendedStart

	for currentEBRPos != Models.EBR_END {
		file.Seek(currentEBRPos, 0)

		var ebr Models.EBR
		binary.Read(file, binary.LittleEndian, &ebr)


		// Si encontramos un EBR con partición lógica válida, agregarlo
		if ebr.PartS > 0 && ebr.GetLogicalPartitionName() != "" {
			logicalPartitions = append(logicalPartitions, ebr)
		}

		// Seguir al siguiente EBR en la cadena
		if ebr.PartNext != Models.EBR_END {
			currentEBRPos = ebr.PartNext
		} else {
			break
		}
	}

	return logicalPartitions, nil
}

// GetBaseGraphConfig retorna la configuración base para gráficos Graphviz
func GetBaseGraphConfig(totalHeight float64) string {
	return fmt.Sprintf(`digraph Report {
    node [shape=plaintext, fontname="Arial"];
    rankdir=TB;
    bgcolor="#2a2a2a";
    dpi=1500;
    size="12,%.1f";
    margin=0;
    ratio=fill;

`, totalHeight)
}

// CalculateGraphHeight calcula la altura dinámica basada en el contenido
func CalculateGraphHeight(partitionCount, logicalCount int) float64 {
	baseHeight := 8.0
	partitionHeight := float64(partitionCount) * 3.5
	logicalHeight := float64(logicalCount) * 2.5
	return baseHeight + partitionHeight + logicalHeight
}

// GetTableHeaderStyle retorna el estilo HTML para headers de tabla
func GetTableHeaderStyle(bgColor, textColor, title string) string {
	return fmt.Sprintf(`            <TR>
                <TD COLSPAN="2" BGCOLOR="%s" ALIGN="center">
                    <FONT COLOR="%s" POINT-SIZE="14"><B>%s</B></FONT>
                </TD>
            </TR>
`, bgColor, textColor, title)
}

// GetTableRowStyle retorna el estilo HTML para filas de tabla
func GetTableRowStyle(label, value string) string {
	return fmt.Sprintf(`                        <TR>
                            <TD BGCOLOR="#2a2a2a" ALIGN="center"><FONT COLOR="#f0f0f0"><B>%s</B></FONT></TD>
                            <TD BGCOLOR="#2a2a2a" ALIGN="center"><FONT COLOR="#f0f0f0">%s</FONT></TD>
                        </TR>
`, label, value)
}

// GetTableWrapperStart retorna el HTML de inicio para tablas
func GetTableWrapperStart() string {
	return `                    <TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" COLOR="#4a4a4a">
`
}

// GetTableWrapperEnd retorna el HTML de cierre para tablas
func GetTableWrapperEnd() string {
	return `                    </TABLE>
`
}

// GetSeparatorRow retorna una fila separadora
func GetSeparatorRow(height string) string {
	return fmt.Sprintf(`            <TR><TD COLSPAN="2" HEIGHT="%s"></TD></TR>
`, height)
}

// FormatPermissions convierte permisos binarios a formato octal legible
func FormatPermissions(perm [3]byte) string {
	return fmt.Sprintf("%o%o%o", perm[0], perm[1], perm[2])
}

// GetFileTypeColor devuelve color apropiado según tipo de archivo
func GetFileTypeColor(fileType int32) string {
	switch fileType {
	case Models.INODO_DIRECTORIO:
		return "lightblue"
	case Models.INODO_ARCHIVO:
		return "lightgreen"
	default:
		return "gray"
	}
}

// GetFileTypeShape devuelve forma apropiada según tipo de archivo
func GetFileTypeShape(fileType int32) string {
	switch fileType {
	case Models.INODO_DIRECTORIO:
		return "folder"
	case Models.INODO_ARCHIVO:
		return "note"
	default:
		return "box"
	}
}

// EscapeGraphvizLabel escapa caracteres especiales para Graphviz
func EscapeGraphvizLabel(text string) string {
	text = strings.ReplaceAll(text, "\"", "\\\"")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	return text
}

// FormatUnixTime convierte timestamp Unix a formato legible
func FormatUnixTime(timestamp float64) string {
	if timestamp == 0 {
		return "N/A"
	}
	// Convertir a formato dd/mm/yyyy hh:mm
	// Para simplificar, usamos formato básico
	return fmt.Sprintf("%.0f", timestamp)
}

// GetUserColor devuelve color único para cada usuario
func GetUserColor(userID int32) string {
	colors := []string{
		"lightcoral", "lightgreen", "lightyellow", "lightpink",
		"lightgray", "lightcyan", "lavender", "mistyrose",
	}

	if userID < int32(len(colors)) {
		return colors[userID]
	}
	return "white"
}

// IsFragmented determina si un inodo tiene bloques fragmentados
func IsFragmented(inodo *Models.Inodo) bool {
	var usedBlocks []int32

	// Recopilar bloques no vacíos
	for _, block := range inodo.I_block {
		if block != -1 {
			usedBlocks = append(usedBlocks, block)
		}
	}

	// Si tiene menos de 2 bloques, no está fragmentado
	if len(usedBlocks) < 2 {
		return false
	}

	// Verificar si los bloques son consecutivos
	for i := 1; i < len(usedBlocks); i++ {
		if usedBlocks[i] != usedBlocks[i-1]+1 {
			return true // Hay un salto, está fragmentado
		}
	}

	return false
}