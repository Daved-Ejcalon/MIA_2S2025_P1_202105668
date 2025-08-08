package main

import (
	"Logica/core"
	"fmt"
	"log"
)

func main() {
	err := core.MkDisk(10, "M", "FF", "./Disco1.mia")
	if err != nil {
		log.Fatal(err)
	}

	err = core.Fdisk(1024, "K", "WF", "./Disco1.mia", "P", "Part1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Prueba completada.")
}
