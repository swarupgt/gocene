package cmd

import (
	"gocene/config"
	"gocene/internal/api"
	"gocene/internal/store"
	"log"
)

func Begin() {
	log.Println("beginning service")

	store.Init()
	config.LoadEnv()

	router := api.GetRouter()
	router.SetEndpoints()
	router.StartRouter()

	log.Println("service started successfully")
}
