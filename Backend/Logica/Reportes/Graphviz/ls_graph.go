package Graphviz

import (
	"MIA_2S2025_P1_202105668/Utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LsEntry representa una entrada de directorio para el reporte ls (redefinición local)
type LsEntry struct {
	Permissions string
	Owner       string
	Group       string
	Size        int32
	Date        string
	Time        string
	Type        string
	Name        string
}

// GenerateLsGraph genera el gráfico DOT para el reporte ls
func GenerateLsGraph(permissions []string, owners []string, groups []string, sizes []int32, dates []string, times []string, types []string, names []string, diskName string, dirPath string, outputPath string) error {
	// Crear entries a partir de los slices
	var entries []LsEntry

	// Verificar que todos los slices tengan la misma longitud
	if len(permissions) != len(owners) || len(owners) != len(groups) || len(groups) != len(sizes) ||
		len(sizes) != len(dates) || len(dates) != len(times) || len(times) != len(types) || len(types) != len(names) {
		return fmt.Errorf("todos los slices deben tener la misma longitud")
	}

	// Construir entries
	for i := 0; i < len(names); i++ {
		entry := LsEntry{
			Permissions: permissions[i],
			Owner:       owners[i],
			Group:       groups[i],
			Size:        sizes[i],
			Date:        dates[i],
			Time:        times[i],
			Type:        types[i],
			Name:        names[i],
		}
		entries = append(entries, entry)
	}
	// Generar contenido DOT
	dotContent := generateLsDotContent(entries, diskName, dirPath)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "ls_report.dot")

	err := os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen JPG usando Graphviz
	return Utils.GenerateImageFromDot(dotFile, outputPath)
}

// generateLsDotContent genera el contenido DOT específico para el reporte ls
func generateLsDotContent(entries []LsEntry, diskName string, dirPath string) string {
	var dot strings.Builder

	// Usar tamaño fijo para evitar deformaciones con números grandes
	totalHeight := 16.0 // Altura fija
	totalWidth := 20.0  // Ancho fijo

	// Configuración base con tamaño fijo
	dot.WriteString(fmt.Sprintf(`digraph Report {
    node [shape=plaintext, fontname="Arial"];
    rankdir=TB;
    bgcolor="#2a2a2a";
    dpi=150;
    size="%.1f,%.1f!";
    fixedsize=true;
    margin=0.2;

`, totalWidth, totalHeight))

	// Crear tabla HTML con diseño consistente
	dot.WriteString("    ls_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#2a2a2a\">\n")

	// Header principal
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"8\" BGCOLOR=\"#5b21b6\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"24\"><B>REPORTE LS</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("10"))

	// Información del disco y directorio
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"8\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#4a4a4a\">\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Disco</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", diskName))
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Directorio</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", dirPath))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("15"))

	// Headers de la tabla de archivos
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"8\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#4a4a4a\">\n")

	// Fila de headers con anchos fijos
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"90\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Permisos</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"60\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Owner</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"60\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Grupo</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"80\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Size(Bytes)</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"80\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Fecha</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"50\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Hora</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"60\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Tipo</B></FONT></TD>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#4a4a4a\" ALIGN=\"center\" WIDTH=\"100\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"10\"><B>Name</B></FONT></TD>\n")
	dot.WriteString("                        </TR>\n")

	// Datos de archivos y carpetas
	if len(entries) == 0 {
		// Mostrar mensaje si el directorio está vacío
		dot.WriteString("                        <TR>\n")
		dot.WriteString("                            <TD COLSPAN=\"8\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><I>(Directorio vacío)</I></FONT></TD>\n")
		dot.WriteString("                        </TR>\n")
	} else {
		// Mostrar cada entrada del directorio
		for _, entry := range entries {
			// Determinar color de fondo basado en tipo
			bgColor := "#2a2a2a"
			if entry.Type == "Carpeta" {
				bgColor = "#1e3a8a" // Azul más oscuro para carpetas
			}

			dot.WriteString("                        <TR>\n")
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"90\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Permissions))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"60\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Owner))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"60\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Group))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"80\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%d</FONT></TD>\n", bgColor, entry.Size))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"80\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Date))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"50\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Time))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"60\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Type))
			dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"%s\" ALIGN=\"center\" WIDTH=\"100\"><FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"9\">%s</FONT></TD>\n", bgColor, entry.Name))
			dot.WriteString("                        </TR>\n")
		}
	}

	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}