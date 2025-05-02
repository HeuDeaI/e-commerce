package skintype

import (
	"net/http"
	"strconv"

	"e-commerce/internal/domains"

	"github.com/gin-gonic/gin"
)

type SkinTypeHandler struct {
	service SkinTypeService
}

func NewSkinTypeHandler(service SkinTypeService) *SkinTypeHandler {
	return &SkinTypeHandler{service: service}
}

func (h *SkinTypeHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/skin-types", h.CreateSkinType)
	router.GET("/skin-types/:id", h.GetSkinTypeByID)
	router.PUT("/skin-types/:id", h.UpdateSkinType)
	router.DELETE("/skin-types/:id", h.DeleteSkinType)
	router.GET("/skin-types", h.GetAllSkinTypes)
}

func (h *SkinTypeHandler) CreateSkinType(c *gin.Context) {
	var skinType domains.SkinType
	if err := c.ShouldBindJSON(&skinType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdSkinType, err := h.service.CreateSkinType(c.Request.Context(), &skinType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdSkinType)
}

func (h *SkinTypeHandler) GetSkinTypeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skin type id"})
		return
	}

	skinType, err := h.service.GetSkinTypeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skinType)
}

func (h *SkinTypeHandler) UpdateSkinType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skin type id"})
		return
	}

	var skinType domains.SkinType
	if err := c.ShouldBindJSON(&skinType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedSkinType, err := h.service.UpdateSkinType(c.Request.Context(), id, &skinType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSkinType)
}

func (h *SkinTypeHandler) DeleteSkinType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skin type id"})
		return
	}

	if err := h.service.DeleteSkinType(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SkinTypeHandler) GetAllSkinTypes(c *gin.Context) {
	skinTypes, err := h.service.GetAllSkinTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skinTypes)
}
