package Reportes

import (
	"fmt"
	"path/filepath"
	"strings"
	"MIA_2S2025_P1_202105668/Logica/Reportes/Graphviz"
)

// ReportGenerator define la interfaz para generadores de reportes
type ReportGenerator interface {
	Generate(partitionID string, outputPath string) error
	ValidateParameters() error
	GetSupportedFormats() []string
}

// ReportType define los tipos de reporte disponibles
type ReportType string

const (
	ReportTypeInode      ReportType = "inode"
	ReportTypeDisk       ReportType = "disk"
	ReportTypeMBR        ReportType = "mbr"
	ReportTypeEBR        ReportType = "ebr"
	ReportTypeSuperBlock ReportType = "sb"
	ReportTypeFile       ReportType = "file"
	ReportTypeLs         ReportType = "ls"
)

// ReportFactory crea instancias de generadores de reportes
type ReportFactory struct{}

// CreateReport crea un generador de reporte según el tipo y formato
func (rf *ReportFactory) CreateReport(reportType ReportType, format string, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Determinar formato basado en la extensión si no se especifica
	if format == "" {
		ext := strings.ToLower(filepath.Ext(outputPath))
		switch ext {
		case ".jpg", ".jpeg":
			format = "jpg"
		case ".png":
			format = "png"
		default:
			format = "jpg" // Por defecto
		}
	}

	switch reportType {
	case ReportTypeInode:
		return rf.createInodeReport(format, outputPath, options)
	case ReportTypeDisk:
		return rf.createDiskReport(format, outputPath, options)
	case ReportTypeMBR:
		return rf.createMBRReport(format, outputPath, options)
	case ReportTypeEBR:
		return rf.createEBRReport(format, outputPath, options)
	case ReportTypeSuperBlock:
		return rf.createSuperBlockReport(format, outputPath, options)
	case ReportTypeFile:
		return rf.createFileReport(format, outputPath, options)
	case ReportTypeLs:
		return rf.createLsReport(format, outputPath, options)
	default:
		return nil, fmt.Errorf("tipo de reporte no soportado: %s", reportType)
	}
}

// createInodeReport crea un generador de reporte de inodos
func (rf *ReportFactory) createInodeReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	if rf.isGraphvizFormat(format) {
		viewType := Graphviz.InodeViewType(options["view"])
		if viewType == "" {
			viewType = Graphviz.ViewTypeStructure // Vista por defecto
		}
		return Graphviz.NewInodeGraphGenerator("", outputPath, format, viewType), nil
	}

	// TODO: Implementar reporte de texto plano
	return nil, fmt.Errorf("formato no soportado para reporte de inodos: %s", format)
}

// createDiskReport crea un generador de reporte de disco
func (rf *ReportFactory) createDiskReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Usar el generador existente
	return &ExistingDiskReportGenerator{outputPath: outputPath}, nil
}

// createMBRReport crea un generador de reporte de MBR
func (rf *ReportFactory) createMBRReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Usar el generador existente
	return &ExistingMBRReportGenerator{outputPath: outputPath}, nil
}

// createEBRReport crea un generador de reporte de EBR
func (rf *ReportFactory) createEBRReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Usar el generador existente
	return &ExistingEBRReportGenerator{outputPath: outputPath}, nil
}

// createSuperBlockReport crea un generador de reporte de superbloque
func (rf *ReportFactory) createSuperBlockReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Usar el generador existente
	return &ExistingSuperBlockReportGenerator{outputPath: outputPath}, nil
}

// createFileReport crea un generador de reporte de archivo
func (rf *ReportFactory) createFileReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Usar el generador existente
	return &ExistingFileReportGenerator{outputPath: outputPath}, nil
}

// createLsReport crea un generador de reporte de listado de directorios
func (rf *ReportFactory) createLsReport(format, outputPath string, options map[string]string) (ReportGenerator, error) {
	// Usar el generador existente
	return &ExistingLsReportGenerator{outputPath: outputPath}, nil
}

// isGraphvizFormat determina si el formato requiere Graphviz
func (rf *ReportFactory) isGraphvizFormat(format string) bool {
	graphvizFormats := []string{"jpg", "jpeg", "png"}
	for _, gf := range graphvizFormats {
		if format == gf {
			return true
		}
	}
	return false
}

// GenerateReport es la función principal que enruta los reportes (mantener compatibilidad)
func GenerateReport(reportName string, partitionID string, outputPath string, pathFileLS string) error {
	switch reportName {
	case "mbr":
		return GenerateMBRReport(partitionID, outputPath)
	case "disk":
		return GenerateDiskReport(partitionID, outputPath)
	case "ebr":
		return GenerateEBRCompleteReport(partitionID, outputPath)
	case "sb":
		return GenerateSuperBlockReport(partitionID, outputPath)
	case "inode":
		// Nueva funcionalidad de reporte de inodos
		factory := &ReportFactory{}
		options := make(map[string]string)
		generator, err := factory.CreateReport(ReportTypeInode, "", outputPath, options)
		if err != nil {
			return err
		}
		return generator.Generate(partitionID, outputPath)
	case "file":
		return GenerateFileReport(partitionID, outputPath, pathFileLS)
	case "ls":
		return GenerateLsReport(partitionID, outputPath, pathFileLS)
	default:
		return fmt.Errorf("tipo de reporte '%s' no reconocido", reportName)
	}
}

// Adaptadores para los generadores existentes
type ExistingDiskReportGenerator struct {
	outputPath string
}

func (e *ExistingDiskReportGenerator) Generate(partitionID string, outputPath string) error {
	return GenerateDiskReport(partitionID, outputPath)
}

func (e *ExistingDiskReportGenerator) ValidateParameters() error {
	return nil
}

func (e *ExistingDiskReportGenerator) GetSupportedFormats() []string {
	return []string{"jpg"}
}

type ExistingMBRReportGenerator struct {
	outputPath string
}

func (e *ExistingMBRReportGenerator) Generate(partitionID string, outputPath string) error {
	return GenerateMBRReport(partitionID, outputPath)
}

func (e *ExistingMBRReportGenerator) ValidateParameters() error {
	return nil
}

func (e *ExistingMBRReportGenerator) GetSupportedFormats() []string {
	return []string{"jpg"}
}

type ExistingEBRReportGenerator struct {
	outputPath string
}

func (e *ExistingEBRReportGenerator) Generate(partitionID string, outputPath string) error {
	return GenerateEBRCompleteReport(partitionID, outputPath)
}

func (e *ExistingEBRReportGenerator) ValidateParameters() error {
	return nil
}

func (e *ExistingEBRReportGenerator) GetSupportedFormats() []string {
	return []string{"jpg"}
}

type ExistingSuperBlockReportGenerator struct {
	outputPath string
}

func (e *ExistingSuperBlockReportGenerator) Generate(partitionID string, outputPath string) error {
	return GenerateSuperBlockReport(partitionID, outputPath)
}

func (e *ExistingSuperBlockReportGenerator) ValidateParameters() error {
	return nil
}

func (e *ExistingSuperBlockReportGenerator) GetSupportedFormats() []string {
	return []string{"jpg"}
}

type ExistingFileReportGenerator struct {
	outputPath string
}

func (e *ExistingFileReportGenerator) Generate(partitionID string, outputPath string) error {
	// El reporte file requiere pathFileLS, pero aquí no lo tenemos disponible
	// Este adaptador no se usa directamente, se maneja en GenerateReport
	return fmt.Errorf("uso directo del adaptador File no soportado, usar GenerateReport")
}

func (e *ExistingFileReportGenerator) ValidateParameters() error {
	return nil
}

func (e *ExistingFileReportGenerator) GetSupportedFormats() []string {
	return []string{"jpg", "png"}
}

type ExistingLsReportGenerator struct {
	outputPath string
}

func (e *ExistingLsReportGenerator) Generate(partitionID string, outputPath string) error {
	// El reporte ls requiere pathFileLS, pero aquí no lo tenemos disponible
	// Este adaptador no se usa directamente, se maneja en GenerateReport
	return fmt.Errorf("uso directo del adaptador Ls no soportado, usar GenerateReport")
}

func (e *ExistingLsReportGenerator) ValidateParameters() error {
	return nil
}

func (e *ExistingLsReportGenerator) GetSupportedFormats() []string {
	return []string{"jpg"}
}