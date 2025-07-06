package controllers

import "github.com/gin-gonic/gin"

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (c *HealthController) Health(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "healthy",
	})
}
