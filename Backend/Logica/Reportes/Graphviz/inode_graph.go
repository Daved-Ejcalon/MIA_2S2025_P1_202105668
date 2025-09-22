package Graphviz

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"MIA_2S2025_P1_202105668/Logica/Disk"
	"MIA_2S2025_P1_202105668/Logica/Users"
	"MIA_2S2025_P1_202105668/Models"
	"MIA_2S2025_P1_202105668/Utils"
)

// InodeGraphGenerator genera reportes gráficos de inodos
type InodeGraphGenerator struct {
	*GraphvizBase
	partitionID string
	mountInfo   *Disk.MountInfo
	superBlock  *Models.SuperBloque
	viewType    InodeViewType
}

// InodeViewType define los tipos de vista disponibles
type InodeViewType string

const (
	ViewTypeStructure     InodeViewType = "structure"     // Jerarquía de archivos
	ViewTypeFragmentation InodeViewType = "fragmentation" // Análisis de fragmentación
	ViewTypeOwnership     InodeViewType = "ownership"     // Agrupado por usuarios
	ViewTypeBitmap        InodeViewType = "bitmap"        // Estado del bitmap
)

// NewInodeGraphGenerator crea un nuevo generador de reportes de inodos
func NewInodeGraphGenerator(partitionID, outputPath, format string, viewType InodeViewType) *InodeGraphGenerator {
	base := NewGraphvizBase("inodos", outputPath, format)
	return &InodeGraphGenerator{
		GraphvizBase: base,
		partitionID:  partitionID,
		viewType:     viewType,
	}
}

// ValidateParameters valida los parámetros del generador
func (ig *InodeGraphGenerator) ValidateParameters() error {
	return nil
}

// GetSupportedFormats retorna los formatos soportados
func (ig *InodeGraphGenerator) GetSupportedFormats() []string {
	return []string{"jpg", "jpeg", "png", "svg", "pdf", "dot"}
}

// Generate genera el reporte de inodos
func (ig *InodeGraphGenerator) Generate(partitionID string, outputPath string) error {
	ig.partitionID = partitionID
	ig.OutputPath = outputPath

	// 1. Cargar datos del sistema
	if err := ig.loadSystemData(); err != nil {
		return fmt.Errorf("error cargando datos del sistema: %v", err)
	}

	// 2. Generar vista de inodos con formato de tabla
	ig.generateInodeTableView()

	// 3. Renderizar
	return ig.SaveAndRender()
}

// loadSystemData carga los datos necesarios del sistema de archivos
func (ig *InodeGraphGenerator) loadSystemData() error {
	var err error
	ig.mountInfo, err = Disk.GetMountInfoByID(ig.partitionID)
	if err != nil {
		return fmt.Errorf("partición no encontrada: %v", err)
	}

	_, ig.superBlock, err = Users.GetPartitionAndSuperBlock(ig.mountInfo)
	if err != nil {
		return fmt.Errorf("error accediendo al sistema de archivos: %v", err)
	}

	return nil
}

// generateStructureView genera vista de estructura de inodos y bloques
func (ig *InodeGraphGenerator) generateStructureView() {
	ig.AddComment("=== VISTA DE ESTRUCTURA DE INODOS ===")

	// Cluster de Inodos
	ig.StartCluster("inodos", "Tabla de Inodos", "filled", "lightyellow")
	ig.generateInodeNodes()
	ig.EndCluster()

	// Cluster de Bloques
	ig.StartCluster("bloques", "Bloques de Datos", "filled", "lightcyan")
	ig.generateBlockNodes()
	ig.EndCluster()

	// Relaciones Inodo → Bloque
	ig.AddComment("=== RELACIONES INODO -> BLOQUE ===")
	ig.generateInodeBlockRelations()
}

// generateInodeNodes genera los nodos de inodos
func (ig *InodeGraphGenerator) generateInodeNodes() {
	inodeBitmap := ig.readInodeBitmap()

	for i := 0; i < int(ig.superBlock.S_inodes_count); i++ {
		if Models.IsBitmapBitSet(inodeBitmap, i) {
			inodo := ig.readInodeFromDisk(i)
			if inodo != nil {
				ig.addInodeTableNode(i, inodo)
			}
		}
	}
}

// generateInodeTableView genera vista de inodos con formato de tabla
func (ig *InodeGraphGenerator) generateInodeTableView() {
	ig.StartGraph("digraph")
	ig.SetRankDir("LR")
	ig.AddRawDOT("    bgcolor=\"#2a2a2a\";\n")
	ig.AddRawDOT("    node [shape=plaintext, fontname=\"Arial\", fontsize=11];\n")

	inodeBitmap := ig.readInodeBitmap()
	var inodeNodes []int

	// Recopilar inodos activos
	for i := 0; i < int(ig.superBlock.S_inodes_count); i++ {
		if Models.IsBitmapBitSet(inodeBitmap, i) {
			inodo := ig.readInodeFromDisk(i)
			if inodo != nil {
				inodeNodes = append(inodeNodes, i)
				ig.addInodeTableNode(i, inodo)
			}
		}
	}

	// Generar conexiones entre inodos
	for i := 0; i < len(inodeNodes)-1; i++ {
		from := fmt.Sprintf("inodo%d", inodeNodes[i])
		to := fmt.Sprintf("inodo%d", inodeNodes[i+1])
		ig.AddEdge(from, to, "", "solid", "#cba6f7")
		ig.AddRawDOT(fmt.Sprintf("    %s -> %s [color=\"#cba6f7\", penwidth=2.5, arrowsize=1.3];\n", from, to))
	}

	ig.EndGraph()
}

// addInodeTableNode agrega un nodo de inodo con formato de tabla HTML
func (ig *InodeGraphGenerator) addInodeTableNode(nodeID int, inodo *Models.Inodo) {
	// Formatear fechas
	atime := ig.formatTimestamp(inodo.I_atime)
	mtime := ig.formatTimestamp(inodo.I_mtime)
	ctime := ig.formatTimestamp(inodo.I_ctime)

	// Crear tabla HTML
	htmlTable := fmt.Sprintf(`
        <TABLE BORDER="1" CELLBORDER="0" CELLSPACING="2" BGCOLOR="#2a2a2a">
            <TR><TD COLSPAN="2" ALIGN="CENTER"><FONT COLOR="#cba6f7"><B>INODO - %d</B></FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_uid</FONT></TD><TD><FONT COLOR="#cba6f7">%d</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_gid</FONT></TD><TD><FONT COLOR="#cba6f7">%d</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_size</FONT></TD><TD><FONT COLOR="#cba6f7">%d</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_atime</FONT></TD><TD><FONT COLOR="#cba6f7">%s</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_ctime</FONT></TD><TD><FONT COLOR="#cba6f7">%s</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_mtime</FONT></TD><TD><FONT COLOR="#cba6f7">%s</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_block_1</FONT></TD><TD><FONT COLOR="#cba6f7">%d</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_block_2</FONT></TD><TD><FONT COLOR="#cba6f7">%d</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_block_3</FONT></TD><TD><FONT COLOR="#cba6f7">%d</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_perm</FONT></TD><TD><FONT COLOR="#cba6f7">%s</FONT></TD></TR>
            <TR><TD ALIGN="LEFT"><FONT COLOR="#000000">i_type</FONT></TD><TD><FONT COLOR="#cba6f7">%s</FONT></TD></TR>
        </TABLE>
    `,
		nodeID,
		inodo.I_uid,
		inodo.I_gid,
		inodo.I_s,
		atime,
		ctime,
		mtime,
		ig.formatBlockNumber(inodo.I_block[0]),
		ig.formatBlockNumber(inodo.I_block[1]),
		ig.formatBlockNumber(inodo.I_block[2]),
		ig.formatPermissions(inodo.I_perm),
		ig.formatInodeType(inodo.I_type),
	)

	ig.AddNodeWithHTML(fmt.Sprintf("inodo%d", nodeID), htmlTable, "plaintext", "none", "transparent")
}

// generateBlockNodes genera los nodos de bloques
func (ig *InodeGraphGenerator) generateBlockNodes() {
	// Leer bitmap de bloques para determinar cuáles están ocupados
	blockBitmap := ig.readBlockBitmap()

	for i := 0; i < int(ig.superBlock.S_blocks_count); i++ {
		if Models.IsBitmapBitSet(blockBitmap, i) {
			ig.addBlockNode(i)
		}
	}
}

// addBlockNode agrega un nodo de bloque al grafo
func (ig *InodeGraphGenerator) addBlockNode(blockID int) {
	label := fmt.Sprintf("Bloque %d", blockID)
	color := "lightgreen"
	shape := "box"

	ig.AddNode(fmt.Sprintf("block%d", blockID), label, shape, "filled", color)
}

// generateInodeBlockRelations genera las relaciones entre inodos y bloques
func (ig *InodeGraphGenerator) generateInodeBlockRelations() {
	inodeBitmap := ig.readInodeBitmap()

	for i := 0; i < int(ig.superBlock.S_inodes_count); i++ {
		if Models.IsBitmapBitSet(inodeBitmap, i) {
			inodo := ig.readInodeFromDisk(i)
			if inodo != nil {
				ig.addInodeBlockEdges(i, inodo)
			}
		}
	}
}

// addInodeBlockEdges agrega las aristas entre un inodo y sus bloques
func (ig *InodeGraphGenerator) addInodeBlockEdges(inodeID int, inodo *Models.Inodo) {
	for j, blockNum := range inodo.I_block {
		if blockNum != -1 {
			fromNode := fmt.Sprintf("inode%d", inodeID)
			toNode := fmt.Sprintf("block%d", blockNum)
			label := fmt.Sprintf("i_block[%d]", j)

			ig.AddEdge(fromNode, toNode, label, "solid", "black")
		}
	}
}

// generateFragmentationView genera vista de análisis de fragmentación
func (ig *InodeGraphGenerator) generateFragmentationView() {
	ig.AddComment("=== VISTA DE FRAGMENTACIÓN ===")

	inodeBitmap := ig.readInodeBitmap()

	// Clasificar inodos por fragmentación
	ig.StartCluster("fragmentados", "Inodos Fragmentados", "filled", "lightcoral")
	ig.StartCluster("consecutivos", "Inodos Consecutivos", "filled", "lightgreen")

	for i := 0; i < int(ig.superBlock.S_inodes_count); i++ {
		if Models.IsBitmapBitSet(inodeBitmap, i) {
			inodo := ig.readInodeFromDisk(i)
			if inodo != nil {
				if Utils.IsFragmented(inodo) {
					ig.addFragmentedInodeNode(i, inodo)
				} else {
					ig.addConsecutiveInodeNode(i, inodo)
				}
			}
		}
	}

	ig.EndCluster() // consecutivos
	ig.EndCluster() // fragmentados
}

// addFragmentedInodeNode agrega un inodo fragmentado
func (ig *InodeGraphGenerator) addFragmentedInodeNode(nodeID int, inodo *Models.Inodo) {
	label := fmt.Sprintf("Inodo %d\\nFRAGMENTADO\\nSize: %d", nodeID, inodo.I_s)
	ig.AddNode(fmt.Sprintf("frag_inode%d", nodeID), label, "box", "filled", "red")
}

// addConsecutiveInodeNode agrega un inodo con bloques consecutivos
func (ig *InodeGraphGenerator) addConsecutiveInodeNode(nodeID int, inodo *Models.Inodo) {
	label := fmt.Sprintf("Inodo %d\\nCONSECUTIVO\\nSize: %d", nodeID, inodo.I_s)
	ig.AddNode(fmt.Sprintf("cons_inode%d", nodeID), label, "box", "filled", "green")
}

// generateOwnershipView genera vista agrupada por usuarios
func (ig *InodeGraphGenerator) generateOwnershipView() {
	ig.AddComment("=== VISTA POR PROPIETARIOS ===")

	userGroups := ig.getUserGroups()

	for userID, inodes := range userGroups {
		clusterName := fmt.Sprintf("user%d", userID)
		clusterLabel := fmt.Sprintf("Usuario UID: %d", userID)
		clusterColor := Utils.GetUserColor(userID)

		ig.StartCluster(clusterName, clusterLabel, "filled", clusterColor)

		for _, inodeID := range inodes {
			inodo := ig.readInodeFromDisk(inodeID)
			if inodo != nil {
				ig.addInodeTableNode(inodeID, inodo)
			}
		}

		ig.EndCluster()
	}
}

// generateBitmapView genera vista del estado del bitmap
func (ig *InodeGraphGenerator) generateBitmapView() {
	ig.AddComment("=== VISTA DEL BITMAP DE INODOS ===")

	inodeBitmap := ig.readInodeBitmap()

	// Crear representación visual del bitmap
	ig.addBitmapVisualization(inodeBitmap)
}

// addBitmapVisualization crea una visualización del bitmap
func (ig *InodeGraphGenerator) addBitmapVisualization(bitmap []byte) {
	// Crear tabla HTML para mostrar el bitmap
	html := `<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0">`
	html += `<TR><TD COLSPAN="8" BGCOLOR="darkblue"><FONT COLOR="white"><B>Bitmap de Inodos</B></FONT></TD></TR>`

	bitIndex := 0
	maxBits := int(ig.superBlock.S_inodes_count)

	for i := 0; i < len(bitmap) && bitIndex < maxBits; i++ {
		html += "<TR>"
		for j := 0; j < 8 && bitIndex < maxBits; j++ {
			bit := (bitmap[i] >> (7 - j)) & 1
			color := "lightgreen" // libre
			if bit == 1 {
				color = "lightcoral" // ocupado
			}
			html += fmt.Sprintf(`<TD BGCOLOR="%s">%d</TD>`, color, bit)
			bitIndex++
		}
		html += "</TR>"

		// Limitar a primeras filas para evitar gráficos muy grandes
		if i >= 7 {
			html += `<TR><TD COLSPAN="8">...</TD></TR>`
			break
		}
	}

	html += "</TABLE>"

	ig.AddNodeWithHTML("bitmap_table", html, "plaintext", "none", "white")
}

// getUserGroups agrupa inodos por usuario propietario
func (ig *InodeGraphGenerator) getUserGroups() map[int32][]int {
	userGroups := make(map[int32][]int)
	inodeBitmap := ig.readInodeBitmap()

	for i := 0; i < int(ig.superBlock.S_inodes_count); i++ {
		if Models.IsBitmapBitSet(inodeBitmap, i) {
			inodo := ig.readInodeFromDisk(i)
			if inodo != nil {
				userGroups[inodo.I_uid] = append(userGroups[inodo.I_uid], i)
			}
		}
	}

	return userGroups
}

// readInodeBitmap lee el bitmap de inodos desde el disco
func (ig *InodeGraphGenerator) readInodeBitmap() []byte {
	file, err := os.Open(ig.mountInfo.DiskPath)
	if err != nil {
		return make([]byte, ig.superBlock.S_inodes_count)
	}
	defer file.Close()

	bitmapPos := ig.getPartitionStart() + int64(ig.superBlock.S_bm_inode_start)
	file.Seek(bitmapPos, 0)

	bitmap := make([]byte, ig.superBlock.S_inodes_count)
	binary.Read(file, binary.LittleEndian, &bitmap)
	return bitmap
}

// readBlockBitmap lee el bitmap de bloques desde el disco
func (ig *InodeGraphGenerator) readBlockBitmap() []byte {
	file, err := os.Open(ig.mountInfo.DiskPath)
	if err != nil {
		return make([]byte, ig.superBlock.S_blocks_count)
	}
	defer file.Close()

	bitmapPos := ig.getPartitionStart() + int64(ig.superBlock.S_bm_block_start)
	file.Seek(bitmapPos, 0)

	bitmap := make([]byte, ig.superBlock.S_blocks_count)
	binary.Read(file, binary.LittleEndian, &bitmap)
	return bitmap
}

// readInodeFromDisk lee un inodo específico desde el disco
func (ig *InodeGraphGenerator) readInodeFromDisk(inodeID int) *Models.Inodo {
	file, err := os.Open(ig.mountInfo.DiskPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	inodePos := ig.getPartitionStart() + int64(ig.superBlock.S_inode_start) + int64(inodeID*Models.INODO_SIZE)
	file.Seek(inodePos, 0)

	var inodo Models.Inodo
	err = binary.Read(file, binary.LittleEndian, &inodo)
	if err != nil {
		return nil
	}

	return &inodo
}

// formatTimestamp formatea un timestamp en formato legible
func (ig *InodeGraphGenerator) formatTimestamp(timestamp float64) string {
	if timestamp == 0 {
		return "00/00/0000 00:00"
	}
	// Convertir timestamp a tiempo Unix y formatear
	t := time.Unix(int64(timestamp), 0)
	return t.Format("02/01/2006 15:04")
}

// formatBlockNumber formatea un número de bloque
func (ig *InodeGraphGenerator) formatBlockNumber(blockNum int32) int32 {
	if blockNum == -1 {
		return -1
	}
	return blockNum
}

// formatPermissions formatea los permisos en formato octal
func (ig *InodeGraphGenerator) formatPermissions(perm [3]byte) string {
	return fmt.Sprintf("%d%d%d", perm[0], perm[1], perm[2])
}

// formatInodeType formatea el tipo de inodo
func (ig *InodeGraphGenerator) formatInodeType(inodeType byte) string {
	if inodeType == Models.INODO_ARCHIVO {
		return "FILE"
	}
	return "DIR"
}

// getPartitionStart obtiene la posición de inicio de la partición
func (ig *InodeGraphGenerator) getPartitionStart() int64 {
	// Esto debería obtener la posición real de la partición desde el MBR
	// Por simplicidad, asumimos que está en el mountInfo
	file, err := os.Open(ig.mountInfo.DiskPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	var mbr Models.MBR
	binary.Read(file, binary.LittleEndian, &mbr)

	// Buscar la partición correspondiente
	for _, partition := range mbr.Partitions {
		if string(partition.PartName[:]) == ig.mountInfo.PartitionName {
			return int64(partition.PartStart)
		}
	}

	return 0
}