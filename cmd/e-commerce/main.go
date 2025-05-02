package main

import (
	"context"
	"time"

	"e-commerce/internal/brand"
	"e-commerce/internal/cache"
	"e-commerce/internal/category"
	"e-commerce/internal/config"
	"e-commerce/internal/database"
	"e-commerce/internal/product"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true)

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load config")
	}

	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	cacheClient, err := cache.New(ctx, &cfg.Cache)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to cache")
	}
	defer cacheClient.Close()

	productRepo := product.NewProductRepository(db.Pool, cacheClient.Client)
	brandRepo := brand.NewBrandRepository(db.Pool, cacheClient.Client)
	categoryRepo := category.NewCategoryRepository(db.Pool, cacheClient.Client)

	productService := product.NewProductService(productRepo)
	brandService := brand.NewBrandService(brandRepo)
	categoryService := category.NewCategoryService(categoryRepo)

	productHandler := product.NewProductHandler(productService)
	brandHandler := brand.NewBrandHandler(brandService)
	categoryHandler := category.NewCategoryHandler(categoryService)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	productHandler.RegisterRoutes(router)
	brandHandler.RegisterRoutes(router)
	categoryHandler.RegisterRoutes(router)

	logrus.Info("Server is running on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		logrus.WithError(err).Fatal("Server failed")
	}
}
