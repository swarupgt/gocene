package cmd

import (
	"gocene/internal/api"
	"gocene/internal/store"
	"log"
)

func Begin() {
	log.Println("beginning service")

	store.Init()

	router := api.GetRouter()
	router.SetEndpoints()
	router.StartRouter()

	log.Println("service started successfully")
}
