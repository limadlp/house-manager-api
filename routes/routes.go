package routes

import (
	"house-manager-api/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(controller *controllers.ListController) *gin.Engine {
	r := gin.Default()

	r.GET("/lists", controller.GetLists)
	r.GET("/lists/:id", controller.GetList)
	r.POST("/lists", controller.CreateList)

	r.POST("/lists/:id/items", controller.AddItem)
	r.PUT("/lists/:id/items/:index", controller.UpdateItem)
	r.DELETE("/lists/:id/items/:index", controller.RemoveItem)

	r.GET("/ws", controller.WebSocketHandler)

	return r
}
