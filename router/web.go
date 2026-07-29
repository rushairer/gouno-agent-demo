package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rushairer/gouno"
	"gouno-agent-demo/internal/httpapi"
)

func RegisterWebRouter(server *gin.Engine, agentHandler *httpapi.Handler) {
	registerWebTestRouter(server)
	registerWebIndexRouter(server)
	if agentHandler != nil {
		server.GET("/healthz", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"status": "ok"}) })
		server.GET("/readyz", agentHandler.Ready)
		v1 := server.Group("/v1/agent")
		v1.POST("/messages", agentHandler.Message)
		v1.POST("/messages/stream", agentHandler.Stream)
	}
}

func registerWebTestRouter(server *gin.Engine) {
	testGroup := server.Group("/test")
	{
		testGroup.GET(
			"/alive",
			func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gouno.NewSuccessResponse("pong"))
			},
		)
	}
}

func registerWebIndexRouter(server *gin.Engine) {
	server.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello gouno!")
	})
}
