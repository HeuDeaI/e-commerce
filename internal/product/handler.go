package product

import (
	"net/http"
	"strconv"
	"strings"

	"e-commerce/internal/domains"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProductHandler interface {
	RegisterRoutes(router *gin.Engine)
}

type productHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) ProductHandler {
	return &productHandler{service: service}
}

func (h *productHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/products", h.createProduct)
	router.GET("/products/:id", h.getProductByID)
	router.PUT("/products/:id", h.updateProduct)
	router.DELETE("/products/:id", h.deleteProduct)
	router.GET("/products", h.getAllProducts)
	router.GET("/products/filter", h.getProductsByFilter)

	router.POST("/products/:id/images", h.uploadProductImage)
	router.DELETE("/products/images/:imageID", h.deleteProductImage)
	router.GET("/products/:id/images", h.getProductImages)
}

func (h *productHandler) createProduct(c *gin.Context) {
	var req domains.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdProduct, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdProduct)
}

func (h *productHandler) getProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *productHandler) updateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req domains.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedProduct, err := h.service.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}

func (h *productHandler) deleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *productHandler) getAllProducts(c *gin.Context) {
	products, err := h.service.GetAllProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *productHandler) getProductsByFilter(c *gin.Context) {
	skinTypeIDs := parseIDs(c.Query("skin-type"))
	brandIDs := parseIDs(c.Query("brand"))
	categoryIDs := parseIDs(c.Query("category"))

	products, err := h.service.GetProductsByFilter(c.Request.Context(), skinTypeIDs, brandIDs, categoryIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

func parseIDs(param string) []int {
	var ids []int
	if param == "" {
		return ids
	}

	strIDs := strings.Split(param, ",")
	for _, strID := range strIDs {
		id, err := strconv.Atoi(strID)
		if err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

func (h *productHandler) uploadProductImage(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	isMain := c.PostForm("is_main") == "true"
	altText := c.PostForm("alt_text")

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		logrus.WithError(err).Error("Failed to open uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}
	defer src.Close()

	image, err := h.service.UploadProductImage(c.Request.Context(), productID, src, isMain, altText)
	if err != nil {
		logrus.WithError(err).Error("Failed to upload product image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusCreated, image)
}

func (h *productHandler) deleteProductImage(c *gin.Context) {
	imageID, err := strconv.Atoi(c.Param("imageID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	if err := h.service.DeleteProductImage(c.Request.Context(), imageID); err != nil {
		logrus.WithError(err).Error("Failed to delete product image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *productHandler) getProductImages(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	images, err := h.service.GetProductImages(c.Request.Context(), productID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get product images")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get images"})
		return
	}

	c.JSON(http.StatusOK, images)
}
