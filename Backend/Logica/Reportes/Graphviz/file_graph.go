package Graphviz

import (
	"MIA_2S2025_P1_202105668/Utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateFileGraph genera el gráfico DOT para el reporte de archivo
func GenerateFileGraph(fileName string, filePath string, content string, diskName string, outputPath string) error {
	// Generar contenido DOT
	dotContent := generateFileGraphDotContent(fileName, filePath, content, diskName)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "file_report.dot")

	err := os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen JPG usando Graphviz
	return Utils.GenerateImageFromDot(dotFile, outputPath)
}

// generateFileGraphDotContent genera el contenido DOT específico para el reporte de archivo
func generateFileGraphDotContent(fileName string, filePath string, content string, diskName string) string {
	var dot strings.Builder

	// Configurar ancho de línea según el tipo de contenido
	lineWidth := 60 // Ancho más corto para archivos con muchos bytes
	if len(content) > 1000 {
		lineWidth = 50 // Aún más corto para archivos muy grandes
	}

	// Calcular altura dinámica basada en contenido tabulado
	tabulatedContent := tabulateContent(content, lineWidth)
	lines := strings.Count(tabulatedContent, "\n") + 1
	contentHeight := float64(lines) * 0.22 // Altura por línea ajustada
	totalHeight := 12.0 + contentHeight   // Base + altura del contenido

	// Limitar altura máxima para evitar imágenes demasiado grandes
	if totalHeight > 25.0 {
		totalHeight = 25.0
	}

	// Usar configuración base
	dot.WriteString(Utils.GetBaseGraphConfig(totalHeight))

	// Crear tabla HTML con diseño consistente
	dot.WriteString("    file_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#2a2a2a\">\n")

	// Header principal
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#5b21b6\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"24\"><B>REPORTE DE ARCHIVO</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("10"))

	// Información del disco y archivo
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#4a4a4a\">\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Disco</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", diskName))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Nombre del archivo</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", fileName))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Ruta completa</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", filePath))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Tamaño</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%d bytes</FONT></TD>\n", len(content)))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("15"))

	// Contenido del archivo
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#4a4a4a\">\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>CONTENIDO</B></FONT></TD>\n")
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"left\">\n")

	// Procesar contenido línea por línea
	if content == "" {
		dot.WriteString("                                <FONT COLOR=\"#f0f0f0\"><I>(El archivo está vacío)</I></FONT>\n")
	} else {
		// Tabular el contenido en líneas más cortas para legibilidad
		tabulatedContent := tabulateContent(content, lineWidth)

		// Escapar caracteres especiales para HTML
		escapedContent := strings.ReplaceAll(tabulatedContent, "&", "&amp;")
		escapedContent = strings.ReplaceAll(escapedContent, "<", "&lt;")
		escapedContent = strings.ReplaceAll(escapedContent, ">", "&gt;")
		escapedContent = strings.ReplaceAll(escapedContent, "\n", "<BR/>")

		dot.WriteString(fmt.Sprintf("                                <FONT COLOR=\"#f0f0f0\" FACE=\"monospace\" POINT-SIZE=\"10\">%s</FONT>\n", escapedContent))
	}

	dot.WriteString("                            </TD>\n")
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}

// tabulateContent divide el contenido en líneas más cortas para mejor legibilidad
func tabulateContent(content string, maxLineLength int) string {
	if content == "" {
		return content
	}

	var result strings.Builder
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Si la línea es más corta que el máximo, mantenerla como está
		if len(line) <= maxLineLength {
			result.WriteString(line)
		} else {
			// Dividir línea larga en múltiples líneas
			for j := 0; j < len(line); j += maxLineLength {
				end := j + maxLineLength
				if end > len(line) {
					end = len(line)
				}

				if j > 0 {
					result.WriteString("\n") // Nueva línea para continuación
				}
				result.WriteString(line[j:end])
			}
		}

		// Agregar salto de línea si no es la última línea
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}