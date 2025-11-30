package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func respondBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func respondNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

func respondInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

func respondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func respondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func respondSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}
