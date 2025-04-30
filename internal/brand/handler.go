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

func (h *BrandHandler) GetAllBrands(c *gin.Context) {
	brands, err := h.service.GetAllBrands(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, brands)
}
