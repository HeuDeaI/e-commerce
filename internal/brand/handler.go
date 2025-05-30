package brand

import (
	"net/http"
	"strconv"

	"e-commerce/internal/domains"

	"github.com/gin-gonic/gin"
)

type BrandHandler struct {
	service BrandService
}

func NewBrandHandler(service BrandService) *BrandHandler {
	return &BrandHandler{service: service}
}

func (h *BrandHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/brands", h.CreateBrand)
	router.GET("/brands/:id", h.GetBrandByID)
	router.PUT("/brands/:id", h.UpdateBrand)
	router.DELETE("/brands/:id", h.DeleteBrand)
	router.GET("/brands", h.GetAllBrands)
}

// @Summary Create a new brand
// @Description Create a new brand with the provided details
// @Tags brands
// @Accept json
// @Produce json
// @Param brand body domains.Brand true "Brand object"
// @Success 201 {object} domains.Brand
// @Failure 400 {object} domains.Error
// @Failure 500 {object} domains.Error
// @Router /brands [post]
func (h *BrandHandler) CreateBrand(c *gin.Context) {
	var brand domains.Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdBrand, err := h.service.CreateBrand(c.Request.Context(), &brand)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdBrand)
}

// @Summary Get brand by ID
// @Description Get a brand by its ID
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Success 200 {object} domains.Brand
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /brands/{id} [get]
func (h *BrandHandler) GetBrandByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand id"})
		return
	}

	brand, err := h.service.GetBrandByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, brand)
}

// @Summary Update brand
// @Description Update an existing brand
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Param brand body domains.Brand true "Brand object"
// @Success 200 {object} domains.Brand
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /brands/{id} [put]
func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand id"})
		return
	}

	var brand domains.Brand
	if err := c.ShouldBindJSON(&brand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedBrand, err := h.service.UpdateBrand(c.Request.Context(), id, &brand)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedBrand)
}

// @Summary Delete brand
// @Description Delete a brand by its ID
// @Tags brands
// @Accept json
// @Produce json
// @Param id path int true "Brand ID"
// @Success 204 "No Content"
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /brands/{id} [delete]
func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid brand id"})
		return
	}

	if err := h.service.DeleteBrand(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Get all brands
// @Description Get a list of all brands
// @Tags brands
// @Accept json
// @Produce json
// @Success 200 {array} domains.Brand
// @Failure 500 {object} domains.Error
// @Router /brands [get]
func (h *BrandHandler) GetAllBrands(c *gin.Context) {
	brands, err := h.service.GetAllBrands(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, brands)
}
