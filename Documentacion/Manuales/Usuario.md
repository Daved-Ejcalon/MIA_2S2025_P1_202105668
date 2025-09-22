# Manual de Usuario – Proyecto I (Go_Disk)
#### Daved Ejcalon Chonay - 202105668 - Lab MIA
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

## Descripción General
**ExtreamFS (GoDisk)** es una aplicación web multiplataforma que simula el funcionamiento de un sistema de archivos **EXT2** sobre discos virtuales, utilizando archivos binarios con extensión `.mia` como contenedores.  

La aplicación está desarrollada en **lenguaje Go** y emplea **Graphviz** para la generación de reportes visuales del sistema de archivos.

---

## Áreas de la ventana principal

### 1. Área de Entrada
- Espacio destinado para el **ingreso de comandos**.  
- El usuario puede escribir cualquier instrucción.  
- Si el comando ingresado es válido, será procesado.  
- Si el comando es incorrecto, aun así se intenta procesar y se notificará el error correspondiente.  

**Ejemplo:**  
![Ventana Principal](https://i.ibb.co/G4YY0BP3/PI-MIA-1.png)

---

### 2. Área de Salida
- Muestra el **resultado** de los comandos ingresados.  
- Puede mostrar:
  - Respuestas exitosas.
  - Mensajes de error en caso de parámetros o comandos inválidos.  

---

## Ejemplo de error al ingresar un comando inválido

Supongamos que el usuario ingresa un comando con un parámetro no soportado (`Parametro x`).  
En este caso, la aplicación mostrará un mensaje de error en el **área de salida**.

**Entrada:**  
![Entrada Invalida](https://i.ibb.co/sJq5cxVY/PI-MIA-2.png)

**Salida:**  
![Error en la salida](https://i.ibb.co/NnpX0J3S/PI-MIA-3.png)

---

## Ejemplo de comando válido

Cuando el usuario ingresa comandos soportados por la aplicación, estos serán procesados correctamente y el área de salida mostrará la respuesta adecuada.

**Ejemplo:**  
![Comando Correcto](https://i.ibb.co/21g1LYGc/PI-MIA-4.png)

---

## Lista de Comandos Permitidos
La aplicación cuenta con una lista definida de **comandos soportados**, los cuales podrán ser ejecutados sin generar error.  


#### Comandos de Administración de Discos
```bash
mkdisk -size=50 -unit=M -fit=FF -path=/ruta/Disco1.mia
rmdisk -path=/ruta/Disco1.mia
fdisk -type=P -unit=M -name=Part1 -size=10 -path=/ruta/Disco1.mia
mount -path=/ruta/Disco1.mia -name=Part1
```

#### Comandos de Reportes
```bash
rep -id=681A -path=/reportes/Disk_681A.jpg -name=disk
```

#### Comandos de Formateo
```bash
mkfs -type=full -id=681A
```

#### Comandos de Gestión de Usuarios y Grupos
```bash
mkgrp -name=usuarios
mkusr -user=Pedro -pass=123 -grp=usuarios
```

#### Comandos de Sesión
```bash
login -user=root -pass=123 -id=681A
logout
```

#### Comandos de Directorios y Archivos
```bash
mkdir -p -path=/home/docs/proyecto
mkfile -path=/home/docs/proyecto/tarea.txt -size=100
cat -file1=/home/docs/proyecto/tarea.txt
```