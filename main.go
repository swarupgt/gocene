package main

import (
	"gocene/cmd"
	"log"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file")
	}

	cmd.Begin()
}
