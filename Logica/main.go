package main

import (
	"Logica/core"
	"log"
)

func main() {
	err := core.MkDisk(10, "M", "FF", "./Disco1.mia")
	if err != nil {
		log.Fatal(err)
	}
}
