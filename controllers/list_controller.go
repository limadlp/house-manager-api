package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"house-manager-api/models"
	"house-manager-api/repositories"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ListController struct {
	Repo         repositories.ListRepository
	Clients      map[*websocket.Conn]bool
	ClientsMutex sync.RWMutex
	FirestoreDB  *firestore.Client
}

func NewListController(repo repositories.ListRepository, firestoreDB *firestore.Client) *ListController {
	controller := &ListController{
		Repo:        repo,
		Clients:     make(map[*websocket.Conn]bool),
		FirestoreDB: firestoreDB,
	}

	// Inicia o listener do Firestore
	go controller.listenToFirestoreChanges()

	return controller
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler gerencia conexões WebSocket
func (c *ListController) WebSocketHandler(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("Erro ao atualizar conexão para WebSocket:", err)
		return
	}
	defer conn.Close()

	// Configura conexão
	conn.SetReadLimit(512)
	c.registerClient(conn)
	defer c.unregisterClient(conn)

	// Configura ping/pong para manter a conexão ativa
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				log.Println("Erro ao enviar ping:", err)
				return
			}
		}
	}()

	// Mantém a conexão aberta
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// listenToFirestoreChanges monitora alterações no Firestore e notifica os clientes WebSocket
func (c *ListController) listenToFirestoreChanges() {
	ctx := context.Background()
	collection := c.FirestoreDB.Collection("listas") // Ajuste o nome da coleção conforme necessário

	// Cria um snapshot listener para a coleção
	snapshot := collection.Snapshots(ctx)
	defer snapshot.Stop()

	for {
		iter, err := snapshot.Next()
		if err != nil {
			log.Println("Erro ao receber snapshot do Firestore:", err)
			time.Sleep(5 * time.Second) // Espera antes de tentar novamente
			continue
		}

		for _, change := range iter.Changes {
			switch change.Kind {
			case firestore.DocumentAdded, firestore.DocumentModified:
				var list models.ShoppingList
				if err := change.Doc.DataTo(&list); err != nil {
					log.Println("Erro ao decodificar lista do Firestore:", err)
					continue
				}
				list.ID = change.Doc.Ref.ID
				c.notifyClients(UpdateAction{
					Type:   "LIST",
					ListID: change.Doc.Ref.ID,
					List:   list,
				})
			case firestore.DocumentRemoved:
				c.notifyClients(UpdateAction{
					Type:   "LIST_DELETE",
					ListID: change.Doc.Ref.ID,
				})
			}
		}
	}
}

// notifyClients envia atualizações para todos os clientes conectados
func (c *ListController) notifyClients(update UpdateAction) {
	message, err := json.Marshal(update)
	if err != nil {
		log.Println("Erro ao serializar atualização:", err)
		return
	}

	c.ClientsMutex.RLock()
	defer c.ClientsMutex.RUnlock()

	for conn := range c.Clients {
		go func(conn *websocket.Conn) {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Erro ao enviar mensagem:", err)
				conn.Close()
				c.unregisterClient(conn)
			}
		}(conn)
	}
}

func (c *ListController) registerClient(conn *websocket.Conn) {
	c.ClientsMutex.Lock()
	defer c.ClientsMutex.Unlock()
	c.Clients[conn] = true
}

func (c *ListController) unregisterClient(conn *websocket.Conn) {
	c.ClientsMutex.Lock()
	defer c.ClientsMutex.Unlock()
	delete(c.Clients, conn)
}

// UpdateAction define a estrutura de uma atualização
type UpdateAction struct {
	Type   string      `json:"type"`             // "LIST_CREATE", "LIST", "LIST_DELETE", "CREATE", "UPDATE", "DELETE"
	ListID string      `json:"listId"`           // ID da lista
	ItemID string      `json:"itemId,omitempty"` // ID do item (se aplicável)
	List   interface{} `json:"list,omitempty"`   // Dados da lista (se aplicável)
	Item   interface{} `json:"item,omitempty"`   // Dados do item (se aplicável)
}

// CreateList cria uma nova lista
func (c *ListController) CreateList(ctx *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Nome inválido"})
		return
	}

	id, err := c.Repo.CreateList(req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar lista"})
		return
	}

	list, err := c.Repo.GetListByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter lista criada"})
		return
	}
	list.ID = id

	// Notifica os clientes sobre a nova lista com objeto completo
	c.notifyClients(UpdateAction{
		Type:   "LIST_CREATE",
		ListID: id,
		List:   list,
	})

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetLists retorna todas as listas
func (c *ListController) GetLists(ctx *gin.Context) {
	lists, err := c.Repo.GetAllLists()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar listas"})
		return
	}
	ctx.JSON(http.StatusOK, lists)
}

// GetList retorna uma lista específica
func (c *ListController) GetList(ctx *gin.Context) {
	id := ctx.Param("id")
	list, err := c.Repo.GetListByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Lista não encontrada"})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// AddItem adiciona um item à lista
func (c *ListController) AddItem(ctx *gin.Context) {
	listID := ctx.Param("id")
	var item models.Item
	if err := ctx.ShouldBindJSON(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Gera um ID único para o item e define a data de criação
	item.ID = uuid.New().String()
	item.Created = time.Now()

	err := c.Repo.AddItem(listID, item)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar item"})
		return
	}

	// Notifica os clientes sobre o novo item
	c.notifyClients(UpdateAction{
		Type:   "CREATE",
		ListID: listID,
		ItemID: item.ID,
		Item:   item,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Item adicionado"})
}

// UpdateItem atualiza um item na lista
func (c *ListController) UpdateItem(ctx *gin.Context) {
	listID := ctx.Param("id")
	index, err := strconv.Atoi(ctx.Param("index"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Índice inválido"})
		return
	}

	var updatedItem models.Item
	if err := ctx.ShouldBindJSON(&updatedItem); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Recupera a lista atual para preservar o ID original do item
	list, err := c.Repo.GetListByID(listID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar lista"})
		return
	}
	if index < 0 || index >= len(list.Items) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Índice fora dos limites"})
		return
	}
	updatedItem.ID = list.Items[index].ID

	err = c.Repo.UpdateItem(listID, index, updatedItem)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar item"})
		return
	}

	// Notifica os clientes sobre a atualização do item
	c.notifyClients(UpdateAction{
		Type:   "UPDATE",
		ListID: listID,
		ItemID: updatedItem.ID,
		Item:   updatedItem,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Item atualizado"})
}

// RemoveItem remove um item da lista
func (c *ListController) RemoveItem(ctx *gin.Context) {
	listID := ctx.Param("id")
	index, err := strconv.Atoi(ctx.Param("index"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Índice inválido"})
		return
	}

	// Recupera a lista para obter o ID do item que será removido
	list, err := c.Repo.GetListByID(listID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar lista"})
		return
	}
	if index < 0 || index >= len(list.Items) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Índice fora dos limites"})
		return
	}
	removedItemID := list.Items[index].ID

	err = c.Repo.RemoveItem(listID, index)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover item"})
		return
	}

	// Notifica os clientes sobre a remoção do item
	c.notifyClients(UpdateAction{
		Type:   "DELETE",
		ListID: listID,
		ItemID: removedItemID,
	})

	ctx.JSON(http.StatusOK, gin.H{"message": "Item removido"})
}
