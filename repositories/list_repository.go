package repositories

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"

	"house-manager-api/config"
	"house-manager-api/models"
)

type ListRepository interface {
	CreateList(name string) (string, error)
	GetAllLists() ([]models.ShoppingList, error)
	GetListByID(id string) (*models.ShoppingList, error)
	AddItem(listID string, item models.Item) error
	UpdateItem(listID string, index int, updatedItem models.Item) error
	RemoveItem(listID string, index int) error
}

type listRepository struct {
	db *firestore.Client
}

func NewListRepository() ListRepository {
	return &listRepository{db: config.FirestoreClient}
}

func (r *listRepository) CreateList(name string) (string, error) {
	ctx := context.Background()
	ref, _, err := r.db.Collection("listas").Add(ctx, map[string]interface{}{
		"name":    name,
		"created": firestore.ServerTimestamp,
		"items":   []interface{}{},
	})
	if err != nil {
		return "", err
	}
	return ref.ID, nil
}

func (r *listRepository) GetAllLists() ([]models.ShoppingList, error) {
	ctx := context.Background()
	snapshot, err := r.db.Collection("listas").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var lists []models.ShoppingList
	for _, doc := range snapshot {
		var list models.ShoppingList
		list.ID = doc.Ref.ID
		if err := doc.DataTo(&list); err != nil {
			log.Println("Erro ao converter lista:", err)
			continue
		}
		lists = append(lists, list)
	}
	return lists, nil
}

func (r *listRepository) GetListByID(id string) (*models.ShoppingList, error) {
	ctx := context.Background()
	doc, err := r.db.Collection("listas").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var list models.ShoppingList
	list.ID = doc.Ref.ID
	if err := doc.DataTo(&list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *listRepository) AddItem(listID string, item models.Item) error {
	ctx := context.Background()
	listRef := r.db.Collection("listas").Doc(listID)

	_, err := listRef.Update(ctx, []firestore.Update{
		{
			Path:  "items",
			Value: firestore.ArrayUnion(item),
		},
	})
	return err
}

func (r *listRepository) UpdateItem(listID string, index int, updatedItem models.Item) error {
	ctx := context.Background()
	listRef := r.db.Collection("listas").Doc(listID)

	doc, err := listRef.Get(ctx)
	if err != nil {
		return err
	}

	var list models.ShoppingList
	if err := doc.DataTo(&list); err != nil {
		return err
	}

	if index < 0 || index >= len(list.Items) {
		return nil
	}

	list.Items[index] = updatedItem

	_, err = listRef.Set(ctx, list)
	return err
}

func (r *listRepository) RemoveItem(listID string, index int) error {
	ctx := context.Background()
	listRef := r.db.Collection("listas").Doc(listID)

	doc, err := listRef.Get(ctx)
	if err != nil {
		return err
	}

	var list models.ShoppingList
	if err := doc.DataTo(&list); err != nil {
		return err
	}

	if index < 0 || index >= len(list.Items) {
		return nil
	}

	list.Items = append(list.Items[:index], list.Items[index+1:]...)

	_, err = listRef.Set(ctx, list)
	return err
}
