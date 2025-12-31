package cmd

import (
	"gocene/config"
	"gocene/internal/api"
	"log"
)

func Begin() {
	log.Println("beginning service")

	config.LoadEnv()

	router := api.GetRouter()
	router.SetEndpoints()
	router.StartRouter()

	log.Println("service started successfully")
}
