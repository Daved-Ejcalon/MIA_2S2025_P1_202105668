# Manual Técnico - Sistema de Archivos EXT2
**Daved Abshalon Ejcalon Chonay - 202105668**
**Proyecto 1 - Manejo e Implementación de Archivos**
**MIA - Segundo Semestre 2025**



## Requisitos del Sistema

### Requisitos Mínimos
- **Sistema Operativo:** Windows 10 / Linux Ubuntu 20.04 o superior  
- **Procesador:** Intel i3 o equivalente  
- **Memoria RAM:** 4 GB  
- **Espacio en disco:** 200 MB libres  
- **Dependencias:**  
  - Go 1.18+  
  - Graphviz instalado y accesible desde la línea de comandos  

### Requisitos Recomendados
- **Sistema Operativo:** Windows 11 / Linux Ubuntu 22.04  
- **Procesador:** Intel i5 o superior  
- **Memoria RAM:** 8 GB o más  
- **Espacio en disco:** 500 MB libres  
- **Dependencias adicionales:**  
  - Conexión a internet estable para actualizaciones.  

---

## 1. Arquitectura General

El proyecto implementa un sistema de archivos **EXT2 simplificado** con gestión de usuarios y generación de reportes. La arquitectura está dividida en módulos especializados:

```
Backend/
├── Models/          # Estructuras de datos
├── Logica/
│   ├── Disk/        # Gestión de discos y particiones
│   ├── System/      # Sistema de archivos EXT2
│   ├── Users/       # Gestión de usuarios y permisos
│   └── Reportes/    # Generación de reportes
└── Utils/           # Utilidades generales
```

---

## 2. Modelos de Datos Principales

### **2.1 MBR (Master Boot Record)**
```go
type MBR struct {
    MbrSize        int32        // Tamaño del MBR
    MbrDate        float64      // Fecha de creación
    MbrDskSignature int32       // Firma del disco
    DskFit         byte         // Algoritmo de ajuste (FF, BF, WF)
    Partitions     [4]Partition // 4 particiones primarias/extendidas
}
```
**Funciones principales:**
- Almacena información del disco y particiones
- Controla algoritmos de ajuste de espacio

### **2.2 SuperBloque - NÚCLEO DEL SISTEMA** 
```go
type SuperBloque struct {
    S_filesystem_type   int32   // Tipo de sistema (2 = EXT2)
    S_inodes_count      int32   // Total de inodos
    S_blocks_count      int32   // Total de bloques
    S_free_blocks_count int32   // Bloques libres
    S_free_inodes_count int32   // Inodos libres
    S_mtime             float64 // Última modificación
    S_mnt_count         int32   // Número de montajes
    S_magic             int32   // Número mágico (0xEF53)
    S_inode_s           int32   // Tamaño de inodo
    S_block_s           int32   // Tamaño de bloque
    S_bm_inode_start    int32   // Inicio bitmap inodos
    S_bm_block_start    int32   // Inicio bitmap bloques
    S_inode_start       int32   // Inicio tabla inodos
    S_block_start       int32   // Inicio área bloques
}
```
**Importancia:** Es el **corazón del sistema EXT2**, controla toda la metadata del filesystem.

### **2.3 Inodo - Gestión de Archivos** 
```go
type Inodo struct {
    I_uid   int32     // ID usuario propietario
    I_gid   int32     // ID grupo propietario
    I_s     int32     // Tamaño del archivo
    I_atime float64   // Último acceso
    I_ctime float64   // Creación
    I_mtime float64   // Última modificación
    I_block [15]int32 // Punteros a bloques (12 directos + 3 indirectos)
    I_type  int32     // Tipo (0=archivo, 1=directorio)
    I_perm  int32     // Permisos (formato Unix)
}
```
**Funciones principales:**
- Almacena metadata de archivos y directorios
- Controla permisos y propiedades

---

## 3. Clases Core del Sistema

### **3.1 EXT2Manager - GESTOR PRINCIPAL** 
**Ubicación:** `Backend/Logica/System/ext2_manager.go`

```go
type EXT2Manager struct {
    mountInfo   *MountInfo
    superblock  *Models.SuperBloque
    diskFile    *os.File
}
```

**Responsabilidades principales:**
- **Inicialización del sistema de archivos**
- **Gestión de bitmaps** (inodos y bloques)
- **Lectura/escritura de inodos**
- **Asignación de espacio libre**

**Métodos críticos:**
- `FormatPartition()` - Formatea partición con EXT2
- `AllocateInode()` - Asigna nuevo inodo
- `AllocateBlock()` - Asigna nuevo bloque
- `ReadInode()` - Lee inodo desde disco

### **3.2 EXT2FileManager - Gestión de Archivos** 
**Ubicación:** `Backend/Logica/System/ext2_files.go`

**Responsabilidades:**
- **Creación y lectura de archivos**
- **Navegación del sistema de archivos**
- **Resolución de rutas**

**Métodos importantes:**
- `CreateFile()` - Crea nuevos archivos
- `ReadFileContent()` - Lee contenido completo
- `findFileInode()` - Busca archivos por ruta

### **3.3 EXT2DirectoryManager - Gestión de Directorios** 
**Ubicación:** `Backend/Logica/System/ext2_directories.go`

**Responsabilidades:**
- **Creación de directorios**
- **Listado de contenido**
- **Gestión de entradas de directorio**

**Métodos importantes:**
- `CreateDirectory()` - Crea directorios
- `ListDirectory()` - Lista contenido
- `addDirectoryEntry()` - Agrega entradas

---

## 4. Gestión de Usuarios y Permisos 

### **4.1 UserManager**
**Ubicación:** `Backend/Logica/Users/user_manager.go`

**Responsabilidades:**
- **Autenticación de usuarios**
- **Gestión de grupos**
- **Control de permisos**

**Funciones principales:**
- `CreateUser()` - Crea nuevos usuarios
- `DeleteUser()` - Elimina usuarios
- `ValidatePermissions()` - Verifica permisos

### **4.2 Sistema de Permisos**
```go
type UserRecord struct {
    ID       int
    Group    string
    Username string
    Password string
}
```

**Características:**
- **Usuarios únicos por sistema**
- **Grupos de usuarios**
- **Permisos estilo Unix (rwx)**

---

## 5. Sistema de Reportes 

### **5.1 Arquitectura de Reportes**
**Ubicación:** `Backend/Logica/Reportes/`

**Tipos implementados:**
- **MBR Report** - Visualiza estructura de particiones
- **Disk Report** - Muestra uso del disco
- **SuperBlock Report** - Información del filesystem
- **Inode Report** - Estructura de inodos
- **File Report** - Contenido de archivos (con tabulación)
- **Ls Report** - Listado de directorios

### **5.2 Generación con Graphviz** 
**Ubicación:** `Backend/Logica/Reportes/Graphviz/`

**Características:**
- **Tablas HTML** para formato profesional
- **Colores temáticos** consistentes
- **Responsive sizing** según contenido
- **Tabulación automática** para archivos grandes

**Ejemplo de uso:**
```bash
rep -id=681A -path="reporte.jpg" -name=sb
rep -id=681A -path="archivo.jpg" -path_file_ls="/archivo.txt" -name=file
```

---

## 6. Flujo de Comandos Principales

### **6.1 Formateo de Partición**
```bash
mkfs -type=ext2 -id=681A
```
**Proceso:**
1. Busca partición por ID
2. Crea SuperBloque
3. Inicializa bitmaps
4. Crea inodo raíz (/)
5. Configura archivos de usuarios

### **6.2 Creación de Archivos**
```bash
mkfile -path="/archivo.txt" -size=100 -id=681A
```
**Proceso:**
1. Valida permisos de usuario
2. Busca directorio padre
3. Asigna inodo y bloques
4. Actualiza entrada de directorio
5. Escribe contenido

### **6.3 Gestión de Usuarios**
```bash
mkusr -user=usuario1 -pwd=123 -grp=grupo1 -id=681A
```
**Proceso:**
1. Valida sesión root
2. Verifica usuario único
3. Actualiza archivo users.txt
4. Sincroniza con disco

---

## 7. Características Técnicas Importantes

### **7.1 Persistencia de Datos**
- **Escritura directa a disco** mediante `binary.Write()`
- **Sincronización inmediata** con `file.Sync()`
- **Gestión de offsets** precisos para estructuras

### **7.2 Gestión de Memoria**
- **Bitmaps en memoria** para acceso rápido
- **Cache de SuperBloque** durante operaciones
- **Lectura bajo demanda** de inodos

### **7.3 Algoritmos de Asignación**
- **First Fit** - Primer espacio disponible
- **Best Fit** - Mejor ajuste de tamaño
- **Worst Fit** - Peor ajuste (máximo espacio)

### **7.4 Limitaciones de la Implementación**
- **EXT2 simplificado** (no journaling)
- **Sin enlaces simbólicos** o hard links
- **Permisos básicos** (sin ACLs)
- **Sin fragmentación** de archivos grandes

---

## 8. Comandos de Compilación y Ejecución

### **8.1 Compilación**
```bash
cd Backend
go build -o mia_system.exe
```

### **8.2 Ejecución**
```bash
# Modo comando individual
./mia_system.exe

# Modo servidor web
./mia_system.exe server
```

### **8.3 Comandos Principales**
```bash
# Gestión de discos
mkdisk -size=1000 -path="disco.dsk"
fdisk -size=100 -path="disco.dsk" -name="Part1"
mount -path="disco.dsk" -name="Part1" -id="681A"

# Sistema de archivos
mkfs -type=ext2 -id=681A
mkdir -path="/home" -id=681A
mkfile -path="/file.txt" -size=100 -id=681A

# Usuarios
mkusr -user=user1 -pwd=123 -grp=group1 -id=681A
login -user=user1 -pwd=123 -id=681A

# Reportes
rep -id=681A -path="reporte.jpg" -name=mbr
rep -id=681A -path="archivo.jpg" -path_file_ls="/file.txt" -name=file
```

---

## 9. Diagrama de Flujo de Trabajo del Sistema

**Diagrama:**  
![Ventana Principal](https://i.ibb.co/NgKfQb00/Diagrama.png)


## 10. Conclusiones

El proyecto implementa exitosamente un **sistema de archivos EXT2 funcional** con:

✅ **Gestión completa de discos y particiones**
✅ **Sistema de archivos EXT2 con metadata**
✅ **Sistema de usuarios y permisos**
✅ **Reportes gráficos profesionales**
✅ **Interfaz de comandos completa**
✅ **Persistencia en disco real**

**Clases más críticas:**
1. **EXT2Manager** - Núcleo del sistema
2. **SuperBloque** - Metadata principal
3. **Sistema de Reportes** - Visualización
4. **UserManager** - Seguridad y permisos

El sistema es **robusto, escalable y completo** para fines académicos y comprensión de sistemas de archivos.