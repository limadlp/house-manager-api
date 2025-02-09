package main

import (
	"log"

	"house-manager-api/config"
	"house-manager-api/controllers"
	"house-manager-api/repositories"
	"house-manager-api/routes"
)

func main() {
	firestoreClient, err := config.InitFirebase()
	if err != nil {
		log.Fatalf("Erro ao inicializar Firebase: %v", err)
	}

	repo := repositories.NewListRepository()
	controller := controllers.NewListController(repo, firestoreClient)

	r := routes.SetupRouter(controller)

	log.Println("Servidor rodando na porta 8080 🚀")
	r.Run(":8080")
}
