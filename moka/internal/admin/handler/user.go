package handler

import (
	"moka/internal/admin/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (u *UserHandler) Demo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "demo",
	})
}