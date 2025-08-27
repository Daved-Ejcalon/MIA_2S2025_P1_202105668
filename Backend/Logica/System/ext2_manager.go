package System

import (
	"MIA_2S2025_P1_202105668/Logica/Partition"
	"MIA_2S2025_P1_202105668/Models"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// MountInfo representa información de una partición montada
type MountInfo struct {
	DiskPath      string // Ruta del disco
	PartitionName string // Nombre de la partición
	MountID       string // ID asignado
	DiskLetter    rune   // Letra asignada al disco
	PartNumber    int    // Número de partición montada
}

// EXT2Manager maneja todas las operaciones del sistema de archivos EXT2
type EXT2Manager struct {
	diskPath      string              // Ruta del archivo de disco
	mountInfo     *MountInfo          // Información de la partición montada
	superBloque   *Models.SuperBloque // SuperBloque cargado en memoria
	inodeBitmap   []byte              // Bitmap de inodos cargado en memoria
	blockBitmap   []byte              // Bitmap de bloques cargado en memoria
}

// NewEXT2Manager crea una nueva instancia del manager de EXT2
func NewEXT2Manager(mountInfo *MountInfo) *EXT2Manager {
	return &EXT2Manager{
		diskPath:  mountInfo.DiskPath,
		mountInfo: mountInfo,
	}
}

// FormatPartition formatea la partición con el sistema de archivos EXT2
func (e *EXT2Manager) FormatPartition() error {
	
	// 1. Calcular estructura del sistema de archivos
	partitionSize, err := e.getPartitionSize()
	if err != nil {
		return fmt.Errorf("error obteniendo tamaño")
	}
	
	// Calcular número de inodos y bloques según especificación EXT2
	// Fórmula: tamaño_particion = sizeOf(superblock) + n + 3*n + n*sizeOf(inodos) + 3*n*sizeOf(block)
	// Despejando n: n = (tamaño_particion - sizeOf(superblock)) / (1 + 3 + sizeOf(inodos) + 3*sizeOf(block))
	
	// Tamaño disponible para estructuras (sin superbloque)
	availableSize := partitionSize - int64(Models.SUPERBLOQUE_SIZE)
	
	// Denominador de la ecuación: 1 (bitmap inodos) + 3 (bitmap bloques) + sizeOf(inodos) + 3*sizeOf(block)
	denominator := int64(1 + 3 + Models.INODO_SIZE + 3*Models.BLOQUE_SIZE)
	
	// Calcular n y aplicar floor
	n := math.Floor(float64(availableSize) / float64(denominator))
	totalInodos := int32(n)
	totalBlocks := int32(3 * n) // El número de bloques es el triple que el número de inodos
	
	fmt.Printf("  📊 Calculando estructura EXT2:\n")
	fmt.Printf("     • Tamaño de partición: %d bytes\n", partitionSize)
	fmt.Printf("     • Valor n calculado: %.0f\n", n)
	fmt.Printf("     • Total inodos (n): %d\n", totalInodos)
	fmt.Printf("     • Total bloques (3*n): %d\n", totalBlocks)
	fmt.Printf("     • Bitmap inodos: %d bytes\n", totalInodos/8+1)
	fmt.Printf("     • Bitmap bloques: %d bytes\n", totalBlocks/8+1)
	fmt.Printf("     • Tabla inodos: %d bytes\n", totalInodos*Models.INODO_SIZE)
	fmt.Printf("     • Tabla bloques: %d bytes\n", totalBlocks*Models.BLOQUE_SIZE)
	
	// 2. Crear SuperBloque
	fmt.Printf("  📝 Creando SuperBloque...\n")
	superBloque := Models.NewSuperBloque(totalInodos, totalBlocks)
	e.superBloque = &superBloque
	
	// 3. Inicializar bitmaps con tamaños calculados
	fmt.Printf("  🎯 Inicializando bitmaps...\n")
	inodeBitmapSize := int(totalInodos/8 + 1) // Suficientes bytes para todos los inodos
	blockBitmapSize := int(totalBlocks/8 + 1) // Suficientes bytes para todos los bloques
	
	e.inodeBitmap = Models.CreateBitmap(inodeBitmapSize)
	e.blockBitmap = Models.CreateBitmap(blockBitmapSize)
	
	// Marcar inodo y bloque raíz como ocupados
	Models.SetBitmapBit(e.inodeBitmap, Models.ROOT_INODE)
	Models.SetBitmapBit(e.blockBitmap, 0) // Bloque 0 para directorio raíz
	
	// 4. Crear inodo raíz
	fmt.Printf("  📁 Creando directorio raíz...\n")
	rootInodo := Models.NewRootInodo()
	rootDirectory := Models.NewRootDirectory()
	
	// 5. Escribir estructuras al disco
	fmt.Printf("  💾 Escribiendo estructuras al disco...\n")
	err = e.writeStructuresToDisk(&rootInodo, &rootDirectory)
	if err != nil {
		return fmt.Errorf("error escribiendo estructuras")
	}
	
	// 6. Crear archivo users.txt
	fmt.Printf("  👥 Creando archivo users.txt...\n")
	err = e.createUsersFile()
	if err != nil {
		return fmt.Errorf("error creando users.txt")
	}
	
	fmt.Printf("✅ Formateo EXT2 completado exitosamente\n")
	fmt.Printf("📋 Verificación de la ecuación EXT2:\n")
	
	// Verificar que la ecuación se cumple
	superBloqueSize := int64(Models.SUPERBLOQUE_SIZE)
	bitmapInodosSize := int64(totalInodos/8 + 1)
	bitmapBloquesSize := int64(totalBlocks/8 + 1) 
	tablaInodosSize := int64(totalInodos) * int64(Models.INODO_SIZE)
	tablaBloquesSize := int64(totalBlocks) * int64(Models.BLOQUE_SIZE)
	
	totalCalculado := superBloqueSize + bitmapInodosSize + bitmapBloquesSize + tablaInodosSize + tablaBloquesSize
	
	fmt.Printf("     • SuperBloque: %d bytes\n", superBloqueSize)
	fmt.Printf("     • Bitmap inodos: %d bytes\n", bitmapInodosSize)
	fmt.Printf("     • Bitmap bloques: %d bytes\n", bitmapBloquesSize) 
	fmt.Printf("     • Tabla inodos: %d bytes\n", tablaInodosSize)
	fmt.Printf("     • Tabla bloques: %d bytes\n", tablaBloquesSize)
	fmt.Printf("     • TOTAL CALCULADO: %d bytes\n", totalCalculado)
	fmt.Printf("     • TAMAÑO PARTICIÓN: %d bytes\n", partitionSize)
	fmt.Printf("     • DIFERENCIA: %d bytes\n", partitionSize-totalCalculado)
	
	return nil
}

// getPartitionSize obtiene el tamaño real de la partición montada desde el MBR/EBR
func (e *EXT2Manager) getPartitionSize() (int64, error) {
	// Abrir archivo del disco
	file, err := os.OpenFile(e.diskPath, os.O_RDONLY, 0644)
	if err != nil {
		return 0, fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return 0, fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Buscar la partición por nombre
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == e.mountInfo.PartitionName {
			// Si es partición primaria
			if partition.PartType == 'P' {
				return partition.PartSize, nil
			}
			// Si es partición extendida, buscar la lógica
			if partition.PartType == 'E' {
				return e.getLogicalPartitionSize(&partition)
			}
		}
	}

	return 0, fmt.Errorf("partición %s no encontrada", e.mountInfo.PartitionName)
}

// writeStructuresToDisk escribe todas las estructuras EXT2 al disco
func (e *EXT2Manager) writeStructuresToDisk(rootInodo *Models.Inodo, rootDir *Models.BloqueCarpeta) error {
	// Abrir archivo del disco
	file, err := os.OpenFile(e.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo disco")
	}
	defer file.Close()
	
	// Obtener posición inicial de la partición
	partitionStart, err := e.getPartitionStartPosition()
	if err != nil {
		return fmt.Errorf("error obteniendo inicio")
	}
	
	// 1. Escribir SuperBloque
	err = e.writeSuperBloque(file, partitionStart)
	if err != nil {
		return fmt.Errorf("error escribiendo SuperBloque")
	}
	
	// 2. Escribir bitmap de inodos
	err = e.writeBitmap(file, partitionStart+int64(e.superBloque.S_bm_inode_start), e.inodeBitmap)
	if err != nil {
		return fmt.Errorf("error escribiendo bitmap inodos")
	}
	
	// 3. Escribir bitmap de bloques
	err = e.writeBitmap(file, partitionStart+int64(e.superBloque.S_bm_block_start), e.blockBitmap)
	if err != nil {
		return fmt.Errorf("error escribiendo bitmap bloques")
	}
	
	// 4. Escribir tabla de inodos (empezando con el inodo raíz)
	err = e.writeInodo(file, partitionStart+int64(e.superBloque.S_inode_start), Models.ROOT_INODE, rootInodo)
	if err != nil {
		return fmt.Errorf("error escribiendo inodo raíz")
	}
	
	// 5. Escribir bloque del directorio raíz
	err = e.writeBloque(file, partitionStart+int64(e.superBloque.S_block_start), 0, rootDir)
	if err != nil {
		return fmt.Errorf("error escribiendo directorio raíz")
	}
	
	return nil
}

// getLogicalPartitionSize obtiene el tamaño de una partición lógica desde el EBR
func (e *EXT2Manager) getLogicalPartitionSize(extendedPartition *Models.Partition) (int64, error) {
	ebrManager := Partition.NewEBRManager(e.diskPath, extendedPartition)
	
	// Obtener lista de particiones lógicas
	logicalPartitions, err := ebrManager.GetLogicalPartitions()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo particiones lógicas: %v", err)
	}
	
	// Buscar la partición por nombre
	for _, logicalPart := range logicalPartitions {
		if logicalPart.Name == e.mountInfo.PartitionName {
			return logicalPart.Size, nil
		}
	}
	
	return 0, fmt.Errorf("partición lógica %s no encontrada", e.mountInfo.PartitionName)
}

// getPartitionStartPosition obtiene la posición real de inicio de la partición montada
func (e *EXT2Manager) getPartitionStartPosition() (int64, error) {
	// Abrir archivo del disco
	file, err := os.OpenFile(e.diskPath, os.O_RDONLY, 0644)
	if err != nil {
		return 0, fmt.Errorf("error abriendo disco: %v", err)
	}
	defer file.Close()

	// Leer MBR
	var mbr Models.MBR
	file.Seek(0, 0)
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return 0, fmt.Errorf("error leyendo MBR: %v", err)
	}

	// Buscar la partición por nombre
	for _, partition := range mbr.Partitions {
		if partition.PartStatus != 0 && partition.GetName() == e.mountInfo.PartitionName {
			// Si es partición primaria
			if partition.PartType == 'P' {
				return partition.PartStart, nil
			}
			// Si es partición extendida, buscar la lógica
			if partition.PartType == 'E' {
				return e.getLogicalPartitionStart(&partition)
			}
		}
	}

	return 0, fmt.Errorf("partición %s no encontrada", e.mountInfo.PartitionName)
}

// getLogicalPartitionStart obtiene la posición de inicio de una partición lógica
func (e *EXT2Manager) getLogicalPartitionStart(extendedPartition *Models.Partition) (int64, error) {
	ebrManager := Partition.NewEBRManager(e.diskPath, extendedPartition)
	
	// Obtener lista de particiones lógicas
	logicalPartitions, err := ebrManager.GetLogicalPartitions()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo particiones lógicas: %v", err)
	}
	
	// Buscar la partición por nombre
	for _, logicalPart := range logicalPartitions {
		if logicalPart.Name == e.mountInfo.PartitionName {
			return logicalPart.Start, nil
		}
	}
	
	return 0, fmt.Errorf("partición lógica %s no encontrada", e.mountInfo.PartitionName)
}

// writeSuperBloque escribe el SuperBloque al disco
func (e *EXT2Manager) writeSuperBloque(file *os.File, position int64) error {
	_, err := file.Seek(position, 0)
	if err != nil {
		return err
	}
	
	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, e.superBloque)
	if err != nil {
		return err
	}
	
	_, err = file.Write(buffer.Bytes())
	return err
}

// writeBitmap escribe un bitmap al disco
func (e *EXT2Manager) writeBitmap(file *os.File, position int64, bitmap []byte) error {
	_, err := file.Seek(position, 0)
	if err != nil {
		return err
	}
	
	_, err = file.Write(bitmap)
	return err
}

// writeInodo escribe un inodo específico al disco
func (e *EXT2Manager) writeInodo(file *os.File, tableStart int64, inodoIndex int32, inodo *Models.Inodo) error {
	position := tableStart + int64(inodoIndex)*int64(Models.INODO_SIZE)
	
	_, err := file.Seek(position, 0)
	if err != nil {
		return err
	}
	
	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, inodo)
	if err != nil {
		return err
	}
	
	_, err = file.Write(buffer.Bytes())
	return err
}

// writeBloque escribe un bloque específico al disco
func (e *EXT2Manager) writeBloque(file *os.File, tableStart int64, bloqueIndex int32, bloque interface{}) error {
	position := tableStart + int64(bloqueIndex)*int64(Models.BLOQUE_SIZE)
	
	_, err := file.Seek(position, 0)
	if err != nil {
		return err
	}
	
	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.LittleEndian, bloque)
	if err != nil {
		return err
	}
	
	_, err = file.Write(buffer.Bytes())
	return err
}

// createUsersFile crea el archivo users.txt en el directorio raíz
func (e *EXT2Manager) createUsersFile() error {
	// Contenido inicial del archivo users.txt
	usersContent := "1,G,root\n1,U,root,root,123\n"
	
	fmt.Printf("     • Contenido users.txt: %s", usersContent)
	
	// TODO: Implementar creación del archivo en el sistema EXT2
	// 1. Encontrar inodo libre
	// 2. Crear inodo para users.txt
	// 3. Asignar bloques para el contenido
	// 4. Escribir contenido a los bloques
	// 5. Actualizar directorio raíz para incluir users.txt
	// 6. Actualizar bitmaps
	
	return nil
}


// LoadEXT2Structures carga las estructuras EXT2 desde el disco
func (e *EXT2Manager) LoadEXT2Structures() error {
	// TODO: Implementar carga de estructuras desde disco
	// 1. Leer SuperBloque
	// 2. Cargar bitmaps
	// 3. Validar integridad
	
	return nil
}