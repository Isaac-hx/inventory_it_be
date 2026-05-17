// This file is hold information about main entry point all packages
// Application entry point
package main

import (
	"fmt"
	"inventory-it/internal/auth"
	"inventory-it/internal/config"
	"inventory-it/internal/database"
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

	//Iniate repository auth
	authRepo := auth.NewRepository(db)
	authUsecase := auth.NewUsecase(authRepo, jwtConfig)
	authHandler := auth.NewHandler(authUsecase)

	//Iniate repository user
	userRepo := user.NewRepository(db)
	userUsecase := user.NewUsecase(userRepo)
	userHandler := user.NewHandler(userUsecase)

	//iniate router
	mux := http.NewServeMux()

	authRoutes := auth.NewRoutes(authHandler, mux, jwtConfig)
	authRoutes.RegisterRoutes()

	userRoutes := user.NewRoutes(userHandler, mux, jwtConfig)
	userRoutes.RegisterRoutes()

	fmt.Println("server running on :2511")
	http.ListenAndServe(":2511", mux)
}
