package saywhat

import (
	"github.com/gin-gonic/gin"

	"clickonetwo.io/whisper/server/internal/middleware"
)

func AddRoutes(r *gin.RouterGroup) {
	r.GET("/settings/:profileId", settingsNotImplemented)
	r.POST("/settings/:profileId", settingsNotImplemented)
	r.PUT("/settings/:profileId", settingsNotImplemented)
}

func settingsNotImplemented(c *gin.Context) {
	profileId := c.Param("profileId")
	method := c.Request.Method
	middleware.CtxLogS(c).Errorw("Not Implemented: %s of settings %s", method, profileId)
	c.JSON(500, gin.H{"status": "error", "error": "not implemented: server is under construction"})
}
