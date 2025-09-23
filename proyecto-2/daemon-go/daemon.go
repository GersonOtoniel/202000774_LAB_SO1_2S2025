package main

import (
	"fmt"
	"os/exec"
)

func main(){
	fmt.Println("Hola mundo")
	cmd := exec.Command("echo", "\"Hola Mundo\"")

	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error al ejecutar el comando: %v\n", err)
		return
	}
}


