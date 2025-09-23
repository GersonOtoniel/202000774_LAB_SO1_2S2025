package main

import (
	"log"
	"os"
	"time"
)

func main(){
	f, err := os.OpenFile("var/log/mydaemon.log", os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetOutput("Daemon iniciado")

	for {
		log.Println("Daemon sigue vivo...")
		time.Sleep(5 * time.Second)
	} 
}
