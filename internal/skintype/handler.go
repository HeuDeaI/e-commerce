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

// @Summary Create a new skin type
// @Description Create a new skin type with the provided details
// @Tags skin-types
// @Accept json
// @Produce json
// @Param skinType body domains.SkinType true "Skin Type object"
// @Success 201 {object} domains.SkinType
// @Failure 400 {object} domains.Error
// @Failure 500 {object} domains.Error
// @Router /skin-types [post]
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

// @Summary Get skin type by ID
// @Description Get a skin type by its ID
// @Tags skin-types
// @Accept json
// @Produce json
// @Param id path int true "Skin Type ID"
// @Success 200 {object} domains.SkinType
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /skin-types/{id} [get]
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

// @Summary Update skin type
// @Description Update an existing skin type
// @Tags skin-types
// @Accept json
// @Produce json
// @Param id path int true "Skin Type ID"
// @Param skinType body domains.SkinType true "Skin Type object"
// @Success 200 {object} domains.SkinType
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /skin-types/{id} [put]
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

// @Summary Delete skin type
// @Description Delete a skin type by its ID
// @Tags skin-types
// @Accept json
// @Produce json
// @Param id path int true "Skin Type ID"
// @Success 204 "No Content"
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /skin-types/{id} [delete]
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

// @Summary Get all skin types
// @Description Get a list of all skin types
// @Tags skin-types
// @Accept json
// @Produce json
// @Success 200 {array} domains.SkinType
// @Failure 500 {object} domains.Error
// @Router /skin-types [get]
func (h *SkinTypeHandler) GetAllSkinTypes(c *gin.Context) {
	skinTypes, err := h.service.GetAllSkinTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, skinTypes)
}
