package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/constant"
	"github.com/lionellc/fusion-gate/internal/domain/user/service"
	"github.com/lionellc/fusion-gate/internal/types"
)

type UserHandler struct {
	authService service.AuthService
}

func NewUserHandler(authService service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req types.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.RegisterResp{ID: user.ID, Email: user.Email, Name: user.Name})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req types.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.LoginResp{Token: token, User: types.UserResp{ID: user.ID, Email: user.Email, Name: user.Name}})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userId := c.GetInt64(constant.HeaderXUserId)

	user, err := h.authService.GetById(c.Request.Context(), userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.ProfileResp{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Balance: user.Balance,
		Role:    user.Role.String(),
	})
}
