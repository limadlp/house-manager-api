package routes

import (
	"house-manager-api/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(controller *controllers.ListController) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/lists", controller.GetLists)
	r.GET("/lists/:id", controller.GetList)
	r.POST("/lists", controller.CreateList)

	r.POST("/lists/:id/items", controller.AddItem)
	r.PUT("/lists/:id/items/:index", controller.UpdateItem)
	r.DELETE("/lists/:id/items/:index", controller.RemoveItem)

	r.GET("/ws", controller.WebSocketHandler)

	return r
}
