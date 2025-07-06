package web

import (
	"heimdall/backend/internal/infra"

	"github.com/gin-gonic/gin"
)

func NewServer(DataStore *infra.DataStore) *gin.Engine {
	router := NewRouter(DataStore)
	engine := router.Setup()

	return engine
}
