package Reportes

import (
	"fmt"
)

// GenerateReport es la función principal que enruta los reportes
func GenerateReport(reportName string, partitionID string, outputPath string, pathFileLS string) error {
	switch reportName {
	case "mbr":
		return GenerateMBRReport(partitionID, outputPath)
	case "disk":
		return GenerateDiskReport(partitionID, outputPath)
	default:
		return fmt.Errorf("tipo de reporte '%s' no reconocido", reportName)
	}
}