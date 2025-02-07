package config

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client

func InitFirebase() {
	ctx := context.Background()
	sa := option.WithCredentialsFile("./config/key.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("Erro ao inicializar Firebase: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Erro ao conectar ao Firestore: %v", err)
	}

	FirestoreClient = client
}
