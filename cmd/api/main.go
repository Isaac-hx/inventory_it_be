// This file is hold information about main entry point all packages
// Application entry point
package main

import (
	"fmt"
	"inventory-it/internal/auth"
	"inventory-it/internal/brands"
	"inventory-it/internal/categories"
	"inventory-it/internal/config"
	"inventory-it/internal/database"
	"inventory-it/internal/departments"
	"inventory-it/internal/pkg"
	"inventory-it/internal/user"
	"net/http"
)

func main() {
	//Load config
	cfg := config.LoadConfig("env.yaml")

	//Initiate database connection
	db := database.InitDatabase(cfg)

	//Initiate superuser seeder
	database.SeedSuperUser(db, cfg)

	//Iniate JWT Config

	jwtConfig := pkg.NewJwt(cfg.Jwt.Secretkey)

	//Iniate domain auth
	authRepo := auth.NewRepository(db)
	authUsecase := auth.NewUsecase(authRepo, jwtConfig)
	authHandler := auth.NewHandler(authUsecase)

	//Iniate domain user
	userRepo := user.NewRepository(db)
	userUsecase := user.NewUsecase(userRepo)
	userHandler := user.NewHandler(userUsecase)

	//Iniate domain departments
	departmentRepo := departments.NewRepository(db)
	departmentUsecase := departments.NewUsecase(departmentRepo)
	departmentHandler := departments.NewHandler(departmentUsecase)

	//Iniate domain brands
	brandRepo := brands.NewBrandRepository(db)
	brandUsecase := brands.NewBrandUsecase(brandRepo)
	brandHandler := brands.NewHandler(brandUsecase)

	//Iniate domain categories
	categoryRepo := categories.NewCategoryRepository(db)
	categoryUsecase := categories.NewCategoryUsecase(categoryRepo)
	categoryHandler := categories.NewCategoryHandler(categoryUsecase)

	//iniate router
	mux := http.NewServeMux()

	//Routes auth
	authRoutes := auth.NewRoutes(authHandler, mux, jwtConfig)
	authRoutes.RegisterRoutes()

	//Routes user
	userRoutes := user.NewRoutes(userHandler, mux, jwtConfig)
	userRoutes.RegisterRoutes()

	//Routes department
	departmentRoutes := departments.NewRoutes(departmentHandler, mux, jwtConfig)
	departmentRoutes.RegisterRoutes()

	//Routes brand
	brandRoutes := brands.NewRoutes(brandHandler, mux, jwtConfig)
	brandRoutes.RegisterRoutes()

	//Routes category
	categoryRoutes := categories.NewRoutes(categoryHandler, mux, jwtConfig)
	categoryRoutes.RegisterRoutes()

	fmt.Println("server running on :2511")
	http.ListenAndServe(":2511", mux)
}
