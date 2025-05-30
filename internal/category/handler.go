package category

import (
	"net/http"
	"strconv"

	"e-commerce/internal/domains"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service CategoryService
}

func NewCategoryHandler(service CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/categories", h.CreateCategory)
	router.GET("/categories/:id", h.GetCategoryByID)
	router.PUT("/categories/:id", h.UpdateCategory)
	router.DELETE("/categories/:id", h.DeleteCategory)
	router.GET("/categories", h.GetAllCategories)
}

// @Summary Create a new category
// @Description Create a new category with the provided details
// @Tags categories
// @Accept json
// @Produce json
// @Param category body domains.Category true "Category object"
// @Success 201 {object} domains.Category
// @Failure 400 {object} domains.Error
// @Failure 500 {object} domains.Error
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var category domains.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdCategory, err := h.service.CreateCategory(c.Request.Context(), &category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// @Summary Get category by ID
// @Description Get a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} domains.Category
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	category, err := h.service.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// @Summary Update category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body domains.Category true "Category object"
// @Success 200 {object} domains.Category
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var category domains.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedCategory, err := h.service.UpdateCategory(c.Request.Context(), id, &category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// @Summary Delete category
// @Description Delete a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Get all categories
// @Description Get a list of all categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} domains.Category
// @Failure 500 {object} domains.Error
// @Router /categories [get]
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}
