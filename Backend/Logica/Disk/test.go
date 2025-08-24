package Disk

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// Helper function to capture stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestMounted(t *testing.T) {
	// Save original state
	originalMounted := mountedPartitions

	tests := []struct {
		name           string
		mountedParts   []MountInfo
		expectedPhrases []string
	}{
		{
			name:         "No mounted partitions",
			mountedParts: []MountInfo{},
			expectedPhrases: []string{
				"No hay particiones montadas en el sistema",
			},
		},
		{
			name: "Single mounted partition",
			mountedParts: []MountInfo{
				{
					MountID:       "681A",
					PartitionName: "Partition1",
					DiskPath:      "/home/user/disk1.dsk",
					DiskLetter:    'A',
					PartNumber:    1,
				},
			},
			expectedPhrases: []string{
				"=== PARTICIONES MONTADAS ===",
				"IDs: 681A",
				"=== DETALLE DE MONTAJES ===",
				"ID: 681A | Partición: Partition1 | Disco: /home/user/disk1.dsk | Letra: A | Número: 1",
				"Total de particiones montadas: 1",
			},
		},
		{
			name: "Multiple mounted partitions",
			mountedParts: []MountInfo{
				{
					MountID:       "681A",
					PartitionName: "Part1",
					DiskPath:      "/home/user/disk1.dsk",
					DiskLetter:    'A',
					PartNumber:    1,
				},
				{
					MountID:       "682B",
					PartitionName: "Part2",
					DiskPath:      "/home/user/disk2.dsk",
					DiskLetter:    'B',
					PartNumber:    2,
				},
			},
			expectedPhrases: []string{
				"=== PARTICIONES MONTADAS ===",
				"IDs: 681A, 682B",
				"=== DETALLE DE MONTAJES ===",
				"Total de particiones montadas: 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test data
			mountedPartitions = tt.mountedParts

			// Capture output
			output := captureOutput(func() {
				Mounted()
			})

			// Verify expected phrases are present
			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(output, phrase) {
					t.Errorf("Expected output to contain '%s', but got:\n%s", phrase, output)
				}
			}
		})
	}

	// Restore original state
	mountedPartitions = originalMounted
}

func TestMountedSimple(t *testing.T) {
	// Save original state
	originalMounted := mountedPartitions

	tests := []struct {
		name           string
		mountedParts   []MountInfo
		expectedOutput string
	}{
		{
			name:           "No mounted partitions",
			mountedParts:   []MountInfo{},
			expectedOutput: "No hay particiones montadas",
		},
		{
			name: "Single partition",
			mountedParts: []MountInfo{
				{MountID: "681A"},
			},
			expectedOutput: "Particiones montadas: 681A",
		},
		{
			name: "Multiple partitions",
			mountedParts: []MountInfo{
				{MountID: "681A"},
				{MountID: "682B"},
				{MountID: "683C"},
			},
			expectedOutput: "Particiones montadas: 681A, 682B, 683C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test data
			mountedPartitions = tt.mountedParts

			// Capture output
			output := captureOutput(func() {
				MountedSimple()
			})

			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', but got: '%s'", tt.expectedOutput, strings.TrimSpace(output))
			}
		})
	}

	// Restore original state
	mountedPartitions = originalMounted
}

func TestMountedTable(t *testing.T) {
	// Save original state
	originalMounted := mountedPartitions

	tests := []struct {
		name           string
		mountedParts   []MountInfo
		expectedPhrases []string
	}{
		{
			name:         "No mounted partitions",
			mountedParts: []MountInfo{},
			expectedPhrases: []string{
				"No hay particiones montadas en el sistema",
			},
		},
		{
			name: "Single partition with table format",
			mountedParts: []MountInfo{
				{
					MountID:       "681A",
					PartitionName: "TestPart",
					DiskPath:      "/home/disk.dsk",
					DiskLetter:    'A',
					PartNumber:    1,
				},
			},
			expectedPhrases: []string{
				"╔══════════════════════════════════════════════════════════════════════╗",
				"║                        PARTICIONES MONTADAS                         ║",
				"║   ID   ║   PARTICIÓN   ║            DISCO              ║ LETRA ║   #   ║",
				"681A",
				"TestPart",
				"/home/disk.dsk",
				"Total: 1 particiones montadas",
			},
		},
		{
			name: "Long names truncation test",
			mountedParts: []MountInfo{
				{
					MountID:       "681A",
					PartitionName: "VeryLongPartitionNameThatShouldBeTruncated",
					DiskPath:      "/very/long/path/to/disk/that/should/be/truncated/disk.dsk",
					DiskLetter:    'A',
					PartNumber:    1,
				},
			},
			expectedPhrases: []string{
				"VeryLongPa...",
				"...cated/disk.dsk",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test data
			mountedPartitions = tt.mountedParts

			// Capture output
			output := captureOutput(func() {
				MountedTable()
			})

			// Verify expected phrases are present
			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(output, phrase) {
					t.Errorf("Expected output to contain '%s', but got:\n%s", phrase, output)
				}
			}
		})
	}

	// Restore original state
	mountedPartitions = originalMounted
}

// Benchmark tests
func BenchmarkMounted(b *testing.B) {
	// Setup test data
	mountedPartitions = []MountInfo{
		{MountID: "681A", PartitionName: "Part1", DiskPath: "/disk1.dsk", DiskLetter: 'A', PartNumber: 1},
		{MountID: "682B", PartitionName: "Part2", DiskPath: "/disk2.dsk", DiskLetter: 'B', PartNumber: 2},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		captureOutput(func() {
			Mounted()
		})
	}
}

func BenchmarkMountedSimple(b *testing.B) {
	// Setup test data
	mountedPartitions = []MountInfo{
		{MountID: "681A"},
		{MountID: "682B"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		captureOutput(func() {
			MountedSimple()
		})
	}
}