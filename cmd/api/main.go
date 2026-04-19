// This file is hold information about main entry point all packages
// Application entry point
package main

import (
	"fmt"
	"inventory-it/internal/auth"
	"inventory-it/internal/config"
	"inventory-it/internal/database"
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

	jwtConfig := auth.NewJwt(cfg.Jwt.Secretkey)

	//Iniate repository user
	userRepo := user.NewRepository(db)
	usecaseRepo := user.NewUsecase(userRepo, jwtConfig)
	userHandler := user.NewHandler(usecaseRepo)

	mux := http.NewServeMux()
	userRoutes := user.NewRoutes(*userHandler)
	userRoutes.RegisterRoutes(mux)

	fmt.Println("server running on :2511")
	http.ListenAndServe(":2511", mux)
}
