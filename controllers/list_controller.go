package controllers

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"house-manager-api/models"
	"house-manager-api/repositories"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ListController struct {
	Repo         repositories.ListRepository
	Clients      map[*websocket.Conn]bool
	ClientsMutex sync.Mutex
}

func NewListController(repo repositories.ListRepository) *ListController {
	return &ListController{
		Repo:    repo,
		Clients: make(map[*websocket.Conn]bool),
	}
}

// WebSocket - Notificações em tempo real
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *ListController) WebSocketHandler(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("Erro ao atualizar conexão para WebSocket:", err)
		return
	}
	defer conn.Close()

	c.ClientsMutex.Lock()
	c.Clients[conn] = true
	c.ClientsMutex.Unlock()

	// Mantém conexão aberta
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			c.ClientsMutex.Lock()
			delete(c.Clients, conn)
			c.ClientsMutex.Unlock()
			break
		}
	}
}

// Notifica os clientes WebSocket sobre mudanças na lista
func (c *ListController) notifyClients(message string) {
	c.ClientsMutex.Lock()
	defer c.ClientsMutex.Unlock()
	for conn := range c.Clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Erro ao enviar mensagem:", err)
			conn.Close()
			delete(c.Clients, conn)
		}
	}
}

// Criar uma nova lista
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

	c.notifyClients("Nova lista criada: " + req.Name)
	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

// Obter todas as listas
func (c *ListController) GetLists(ctx *gin.Context) {
	lists, err := c.Repo.GetAllLists()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar listas"})
		return
	}
	ctx.JSON(http.StatusOK, lists)
}

// Obter uma lista específica
func (c *ListController) GetList(ctx *gin.Context) {
	id := ctx.Param("id")
	list, err := c.Repo.GetListByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Lista não encontrada"})
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// Adicionar item à lista
func (c *ListController) AddItem(ctx *gin.Context) {
	id := ctx.Param("id")
	var item models.Item
	if err := ctx.ShouldBindJSON(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	err := c.Repo.AddItem(id, item)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar item"})
		return
	}

	c.notifyClients("Novo item adicionado: " + item.Item)
	ctx.JSON(http.StatusOK, gin.H{"message": "Item adicionado"})
}

// Atualizar item na lista
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

	err = c.Repo.UpdateItem(listID, index, updatedItem)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar item"})
		return
	}

	c.notifyClients("Item atualizado: " + updatedItem.Item)
	ctx.JSON(http.StatusOK, gin.H{"message": "Item atualizado"})
}

// Remover item da lista
func (c *ListController) RemoveItem(ctx *gin.Context) {
	listID := ctx.Param("id")
	index, err := strconv.Atoi(ctx.Param("index"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Índice inválido"})
		return
	}

	err = c.Repo.RemoveItem(listID, index)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover item"})
		return
	}

	c.notifyClients("Item removido")
	ctx.JSON(http.StatusOK, gin.H{"message": "Item removido"})
}
