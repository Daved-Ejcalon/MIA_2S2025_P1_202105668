package Graphviz

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GraphvizBase es la clase base para generar reportes con Graphviz
type GraphvizBase struct {
	Title      string
	OutputPath string
	Format     string // "dot", "jpg", "png", "svg", "pdf"
	RankDir    string // "TB", "LR", "BT", "RL"
	buffer     strings.Builder
}

// NewGraphvizBase crea una nueva instancia base de Graphviz
func NewGraphvizBase(title, outputPath, format string) *GraphvizBase {
	return &GraphvizBase{
		Title:      title,
		OutputPath: outputPath,
		Format:     format,
		RankDir:    "TB",
	}
}

// StartGraph inicia la definición del grafo
func (g *GraphvizBase) StartGraph(graphType string) {
	g.buffer.WriteString(fmt.Sprintf("%s %s {\n", graphType, g.Title))
	g.buffer.WriteString(fmt.Sprintf("    rankdir=%s;\n", g.RankDir))
	g.buffer.WriteString("    node [fontname=\"Arial\" fontsize=10];\n")
	g.buffer.WriteString("    edge [fontname=\"Arial\" fontsize=8];\n")
	g.buffer.WriteString("    bgcolor=\"#2a2a2a\";\n")
	g.buffer.WriteString("    dpi=300;\n\n")
}

// EndGraph cierra la definición del grafo
func (g *GraphvizBase) EndGraph() {
	g.buffer.WriteString("}\n")
}

// AddNode agrega un nodo al grafo
func (g *GraphvizBase) AddNode(id, label, shape, style, color string) {
	escapedLabel := strings.ReplaceAll(label, "\"", "\\\"")
	escapedLabel = strings.ReplaceAll(escapedLabel, "\n", "\\n")

	g.buffer.WriteString(fmt.Sprintf(
		"    %s [label=\"%s\" shape=%s style=%s fillcolor=%s];\n",
		id, escapedLabel, shape, style, color))
}

// AddNodeWithHTML agrega un nodo con formato HTML
func (g *GraphvizBase) AddNodeWithHTML(id, htmlLabel, shape, style, color string) {
	g.buffer.WriteString(fmt.Sprintf(
		"    %s [label=<%s> shape=%s style=%s fillcolor=%s];\n",
		id, htmlLabel, shape, style, color))
}

// AddEdge agrega una arista entre dos nodos
func (g *GraphvizBase) AddEdge(from, to, label, style, color string) {
	if label != "" {
		g.buffer.WriteString(fmt.Sprintf(
			"    %s -> %s [label=\"%s\" style=%s color=%s];\n",
			from, to, label, style, color))
	} else {
		g.buffer.WriteString(fmt.Sprintf(
			"    %s -> %s [style=%s color=%s];\n",
			from, to, style, color))
	}
}

// StartCluster inicia un cluster (subgrafo)
func (g *GraphvizBase) StartCluster(name, label, style, color string) {
	g.buffer.WriteString(fmt.Sprintf("    subgraph cluster_%s {\n", name))
	g.buffer.WriteString(fmt.Sprintf("        label=\"%s\";\n", label))
	g.buffer.WriteString(fmt.Sprintf("        style=%s;\n", style))
	g.buffer.WriteString(fmt.Sprintf("        fillcolor=%s;\n", color))
	g.buffer.WriteString("        fontsize=12;\n")
	g.buffer.WriteString("        fontname=\"Arial\";\n\n")
}

// EndCluster cierra un cluster
func (g *GraphvizBase) EndCluster() {
	g.buffer.WriteString("    }\n\n")
}

// AddComment agrega un comentario al código DOT
func (g *GraphvizBase) AddComment(comment string) {
	g.buffer.WriteString(fmt.Sprintf("    // %s\n", comment))
}

// AddRawDOT agrega código DOT sin procesar
func (g *GraphvizBase) AddRawDOT(dotCode string) {
	g.buffer.WriteString(dotCode)
}

// GetDOTContent retorna el contenido DOT generado
func (g *GraphvizBase) GetDOTContent() string {
	return g.buffer.String()
}

// SaveAndRender guarda el archivo DOT y lo renderiza si es necesario
func (g *GraphvizBase) SaveAndRender() error {
	// Asegurarse de que el directorio existe
	outputDir := filepath.Dir(g.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio: %v", err)
	}

	// Guardar archivo .dot
	dotPath := strings.Replace(g.OutputPath, filepath.Ext(g.OutputPath), ".dot", 1)
	err := os.WriteFile(dotPath, []byte(g.buffer.String()), 0644)
	if err != nil {
		return fmt.Errorf("error guardando archivo DOT: %v", err)
	}

	// Renderizar si no es formato dot
	if g.Format != "dot" {
		return g.renderWithGraphviz(dotPath)
	}
	return nil
}

// renderWithGraphviz ejecuta Graphviz para generar la imagen
func (g *GraphvizBase) renderWithGraphviz(dotPath string) error {
	// Verificar si Graphviz está disponible
	if _, err := exec.LookPath("dot"); err != nil {
		return fmt.Errorf("Graphviz no está instalado o no está en PATH: %v", err)
	}

	// Determinar formato de salida
	format := g.Format
	if format == "jpg" || format == "jpeg" {
		format = "jpg"
	}

	// Ejecutar comando dot con configuración para márgenes mínimos
	cmd := exec.Command("dot", fmt.Sprintf("-T%s", format), "-Gmargin=0", "-Gpad=0", dotPath, "-o", g.OutputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error ejecutando Graphviz: %v\nOutput: %s", err, string(output))
	}

	// Verificar que el archivo se creó
	if _, err := os.Stat(g.OutputPath); os.IsNotExist(err) {
		return fmt.Errorf("el archivo de salida no se generó: %s", g.OutputPath)
	}

	return nil
}

// SetRankDir establece la dirección del grafo
func (g *GraphvizBase) SetRankDir(direction string) {
	g.RankDir = direction
}

// Clear limpia el buffer del grafo
func (g *GraphvizBase) Clear() {
	g.buffer.Reset()
}