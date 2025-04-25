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

	_, err = cache.New(ctx, &cfg.Cache)
	if err != nil {
		log.Fatalf("Failed to connect to cache: %v", err)
	}

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(ctx, "internal/database/migrations"); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	productRepo := product.NewProductRepository(db.Pool)
	brandRepo := brand.NewBrandRepository(db.Pool)
	categoryRepo := category.NewCategoryRepository(db.Pool)

	productService := product.NewProductService(productRepo)
	brandService := brand.NewBrandService(brandRepo)
	categoryService := category.NewCategoryService(categoryRepo)

	productHandler := product.NewProductHandler(productService)
	brandHandler := brand.NewBrandHandler(brandService)
	categoryHandler := category.NewCategoryHandler(categoryService)

	router := gin.Default()

	router.LoadHTMLGlob("web/templates/**/*")

	productHandler.RegisterRoutes(router)
	brandHandler.RegisterRoutes(router)
	categoryHandler.RegisterRoutes(router)

	log.Println("Server is running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
