package main

import (
	"context"
	"log"

	"e-commerce/internal/brand"
	"e-commerce/internal/cache"
	"e-commerce/internal/category"
	"e-commerce/internal/config"
	"e-commerce/internal/database"
	"e-commerce/internal/product"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	cacheClient, err := cache.New(ctx, &cfg.Cache)
	if err != nil {
		log.Fatalf("Failed to connect to cache: %v", err)
	}
	defer cacheClient.Close()

	productRepo := product.NewProductRepository(db.Pool)
	brandRepo := brand.NewBrandRepository(db.Pool)
	categoryRepo := category.NewCategoryRepository(db.Pool)

	productCachedRepo := product.NewCachedProductRepository(cacheClient.Client)
	brandCachedRepo := brand.NewCachedBrandRepository(cacheClient.Client)
	categoryCachedRepo := category.NewCachedCategoryRepository(cacheClient.Client)

	productService := product.NewProductService(productRepo, productCachedRepo)
	brandService := brand.NewBrandService(brandRepo, brandCachedRepo)
	categoryService := category.NewCategoryService(categoryRepo, categoryCachedRepo)

	productHandler := product.NewProductHandler(productService)
	brandHandler := brand.NewBrandHandler(brandService)
	categoryHandler := category.NewCategoryHandler(categoryService)

	router := gin.Default()

	productHandler.RegisterRoutes(router)
	brandHandler.RegisterRoutes(router)
	categoryHandler.RegisterRoutes(router)

	log.Println("Server is running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
