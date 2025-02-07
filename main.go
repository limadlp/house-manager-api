package main

import (
	"log"

	"house-manager-api/config"
	"house-manager-api/controllers"
	"house-manager-api/repositories"
	"house-manager-api/routes"
)

func main() {
	config.InitFirebase()

	repo := repositories.NewListRepository()
	controller := controllers.NewListController(repo)

	r := routes.SetupRouter(controller)

	log.Println("Servidor rodando na porta 8080 🚀")
	r.Run(":8080")
}
