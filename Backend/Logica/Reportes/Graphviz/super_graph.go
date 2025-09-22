package Graphviz

import (
	"MIA_2S2025_P1_202105668/Models"
	"MIA_2S2025_P1_202105668/Utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateSuperBlockGraph genera el gráfico DOT para el reporte del superbloque
func GenerateSuperBlockGraph(superblock *Models.SuperBloque, diskName string, outputPath string) error {
	// Generar contenido DOT
	dotContent := generateSuperBlockDotContent(superblock, diskName)

	// Crear archivo temporal DOT
	tempDir := os.TempDir()
	dotFile := filepath.Join(tempDir, "superblock_report.dot")

	err := os.WriteFile(dotFile, []byte(dotContent), 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer os.Remove(dotFile)

	// Generar imagen JPG usando Graphviz
	return Utils.GenerateImageFromDot(dotFile, outputPath)
}

// generateSuperBlockDotContent genera el contenido DOT específico para el superbloque
func generateSuperBlockDotContent(sb *Models.SuperBloque, diskName string) string {
	var dot strings.Builder

	// Usar configuración base con altura fija para SuperBlock
	totalHeight := 12.0
	dot.WriteString(Utils.GetBaseGraphConfig(totalHeight))

	// Crear tabla HTML con diseño consistente
	dot.WriteString("    sb_table [label=<\n")
	dot.WriteString("        <TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"4\" BGCOLOR=\"#2a2a2a\">\n")

	// Header principal
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#5b21b6\" ALIGN=\"center\">\n")
	dot.WriteString("                    <FONT COLOR=\"#f0f0f0\" POINT-SIZE=\"24\"><B>REPORTE DE SUPERBLOQUE</B></FONT>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("10"))

	// Información del disco
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString("                    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" COLOR=\"#4a4a4a\">\n")
	dot.WriteString("                        <TR>\n")
	dot.WriteString("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\"><B>Disco</B></FONT></TD>\n")
	dot.WriteString(fmt.Sprintf("                            <TD BGCOLOR=\"#2a2a2a\" ALIGN=\"center\"><FONT COLOR=\"#f0f0f0\">%s</FONT></TD>\n", diskName))
	dot.WriteString("                        </TR>\n")
	dot.WriteString("                    </TABLE>\n")
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	// Espacio separador
	dot.WriteString(Utils.GetSeparatorRow("15"))

	// Datos del SuperBloque
	dot.WriteString("            <TR>\n")
	dot.WriteString("                <TD COLSPAN=\"2\" BGCOLOR=\"#2a2a2a\" ALIGN=\"center\">\n")
	dot.WriteString(Utils.GetTableWrapperStart())

	// sb_nombre_hd
	dot.WriteString(Utils.GetTableRowStyle("sb_nombre_hd", diskName))

	// sb_filesystem_type
	dot.WriteString(Utils.GetTableRowStyle("sb_filesystem_type", fmt.Sprintf("%d", sb.S_filesystem_type)))

	// sb_arbol_virtual_count (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_arbol_virtual_count", "0"))

	// sb_detalle_directorio_count (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_detalle_directorio_count", "0"))

	// sb_inodos_count
	dot.WriteString(Utils.GetTableRowStyle("sb_inodos_count", fmt.Sprintf("%d", sb.S_inodes_count)))

	// sb_bloques_count
	dot.WriteString(Utils.GetTableRowStyle("sb_bloques_count", fmt.Sprintf("%d", sb.S_blocks_count)))

	// sb_arbol_virtual_free (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_arbol_virtual_free", "0"))

	// sb_detalle_directorio_free (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_detalle_directorio_free", "0"))

	// sb_inodos_free
	dot.WriteString(Utils.GetTableRowStyle("sb_inodos_free", fmt.Sprintf("%d", sb.S_free_inodes_count)))

	// sb_bloques_free
	dot.WriteString(Utils.GetTableRowStyle("sb_bloques_free", fmt.Sprintf("%d", sb.S_free_blocks_count)))

	// sb_date_creacion
	mtime := "N/A"
	if sb.S_mtime > 0 {
		mtime = time.Unix(int64(sb.S_mtime), 0).Format("2006-01-02 15:04:05")
	}
	dot.WriteString(Utils.GetTableRowStyle("sb_date_creacion", mtime))

	// sb_date_ultimo_montaje
	umtime := "N/A"
	if sb.S_umtime > 0 {
		umtime = time.Unix(int64(sb.S_umtime), 0).Format("2006-01-02 15:04:05")
	}
	dot.WriteString(Utils.GetTableRowStyle("sb_date_ultimo_montaje", umtime))

	// sb_montajes_count
	dot.WriteString(Utils.GetTableRowStyle("sb_montajes_count", fmt.Sprintf("%d", sb.S_mnt_count)))

	// sb_ap_bitmap_arbol_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_bitmap_arbol_directorio", "0"))

	// sb_ap_arbol_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_arbol_directorio", "0"))

	// sb_ap_bitmap_detalle_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_bitmap_detalle_directorio", "0"))

	// sb_ap_detalle_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_detalle_directorio", "0"))

	// sb_ap_bitmap_inodos
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_bitmap_inodos", fmt.Sprintf("%d", sb.S_bm_inode_start)))

	// sb_ap_inodos
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_inodos", fmt.Sprintf("%d", sb.S_inode_start)))

	// sb_ap_bitmap_bloques
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_bitmap_bloques", fmt.Sprintf("%d", sb.S_bm_block_start)))

	// sb_ap_bloques
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_bloques", fmt.Sprintf("%d", sb.S_block_start)))

	// sb_ap_log (no existe en modelo actual, calcular posición aproximada)
	logPos := sb.S_block_start + (sb.S_blocks_count * sb.S_block_s)
	dot.WriteString(Utils.GetTableRowStyle("sb_ap_log", fmt.Sprintf("%d", logPos)))

	// sb_size_struct_arbol_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_size_struct_arbol_directorio", "0"))

	// sb_size_struct_detalle_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_size_struct_detalle_directorio", "0"))

	// sb_size_struct_inodo
	dot.WriteString(Utils.GetTableRowStyle("sb_size_struct_inodo", fmt.Sprintf("%d", sb.S_inode_s)))

	// sb_size_struct_bloque
	dot.WriteString(Utils.GetTableRowStyle("sb_size_struct_bloque", fmt.Sprintf("%d", sb.S_block_s)))

	// sb_first_free_bit_arbol_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_first_free_bit_arbol_directorio", "0"))

	// sb_first_free_bit_detalle_directorio (no existe en modelo actual, usar 0)
	dot.WriteString(Utils.GetTableRowStyle("sb_first_free_bit_detalle_directorio", "0"))

	// sb_first_free_bit_tabla_inodos
	dot.WriteString(Utils.GetTableRowStyle("sb_first_free_bit_tabla_inodos", fmt.Sprintf("%d", sb.S_firts_ino)))

	// sb_first_free_bit_bloques
	dot.WriteString(Utils.GetTableRowStyle("sb_first_free_bit_bloques", fmt.Sprintf("%d", sb.S_first_blo)))

	// sb_magic_num
	dot.WriteString(Utils.GetTableRowStyle("sb_magic_num", fmt.Sprintf("%d", sb.S_magic)))

	dot.WriteString(Utils.GetTableWrapperEnd())
	dot.WriteString("                </TD>\n")
	dot.WriteString("            </TR>\n")

	dot.WriteString("        </TABLE>\n")
	dot.WriteString("    >];\n")
	dot.WriteString("}\n")

	return dot.String()
}