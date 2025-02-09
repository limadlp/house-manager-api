package config

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client

func InitFirebase() (*firestore.Client, error) {
	ctx := context.Background()

	// Configura o Firebase
	opt := option.WithCredentialsFile("./config/key.json") // Substitua pelo caminho correto
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	// Inicializa o Firestore
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	FirestoreClient = client
	log.Println("Firestore inicializado com sucesso")
	return client, nil
}
