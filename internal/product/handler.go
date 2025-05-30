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

// @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param product body domains.ProductRequest true "Product object"
// @Success 201 {object} domains.ProductResponse
// @Failure 400 {object} domains.Error
// @Failure 500 {object} domains.Error
// @Router /products [post]
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

// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} domains.ProductResponse
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /products/{id} [get]
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

// @Summary Update product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body domains.ProductRequest true "Product object"
// @Success 200 {object} domains.ProductResponse
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /products/{id} [put]
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

// @Summary Delete product
// @Description Delete a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /products/{id} [delete]
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

// @Summary Get all products
// @Description Get a list of all products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {array} domains.ProductResponse
// @Failure 500 {object} domains.Error
// @Router /products [get]
func (h *productHandler) getAllProducts(c *gin.Context) {
	products, err := h.service.GetAllProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// @Summary Filter products
// @Description Get products filtered by various criteria
// @Tags products
// @Accept json
// @Produce json
// @Param skin-type query string false "Comma-separated list of skin type IDs"
// @Param brand query string false "Comma-separated list of brand IDs"
// @Param category query string false "Comma-separated list of category IDs"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Success 200 {array} domains.ProductResponse
// @Failure 500 {object} domains.Error
// @Router /products/filter [get]
func (h *productHandler) getProductsByFilter(c *gin.Context) {
	skinTypeIDs := parseIDs(c.Query("skin-type"))
	brandIDs := parseIDs(c.Query("brand"))
	categoryIDs := parseIDs(c.Query("category"))

	var priceRange domains.PriceRange
	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			priceRange.MinPrice = &price
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			priceRange.MaxPrice = &price
		}
	}

	products, err := h.service.GetProductsByFilter(c.Request.Context(), skinTypeIDs, brandIDs, categoryIDs, &priceRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// @Summary Upload product image
// @Description Upload an image for a product
// @Tags products
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Product ID"
// @Param image formData file true "Image file"
// @Param is_main formData bool false "Whether this is the main image"
// @Param alt_text formData string false "Alternative text for the image"
// @Success 201 {object} domains.ProductImage
// @Failure 400 {object} domains.Error
// @Failure 500 {object} domains.Error
// @Router /products/{id}/images [post]
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

// @Summary Delete product image
// @Description Delete a product image by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param imageID path int true "Image ID"
// @Success 204 "No Content"
// @Failure 400 {object} domains.Error
// @Failure 404 {object} domains.Error
// @Router /products/images/{imageID} [delete]
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

// @Summary Get product images
// @Description Get all images for a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {array} domains.ProductImage
// @Failure 400 {object} domains.Error
// @Failure 500 {object} domains.Error
// @Router /products/{id}/images [get]
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
