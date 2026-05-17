//This file contain connection with database

package database

import (
	"database/sql"
	"fmt"
	"inventory-it/internal/config"
	"inventory-it/internal/pkg"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// Initialization database
func InitDatabase(cfg *config.Config) *sql.DB {

	//Config data source name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Dbname,
	)

	//Open connection to driver mysql
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	//Notification success connection
	log.Println("Connection to database success!")
	return db
}

// Seeder superuser
func SeedSuperUser(db *sql.DB, cfg *config.Config) {
	var username string

	err := db.QueryRow("SELECT username FROM users WHERE role = 'superuser'").Scan(&username)
	if err == sql.ErrNoRows {
		queryInsert := "INSERT INTO users(user_id,username,password,email,role,department_id) VALUES(?,?,?,?,?,?)"
		_, err := db.Exec(queryInsert,
			uuid.NewString(),
			cfg.SuperAdmin.Username,
			pkg.NewHashingPassword(cfg.SuperAdmin.Password),
			cfg.SuperAdmin.Email,
			"superuser",
			cfg.SuperAdmin.DepartmentId,
		)

		if err != nil {
			log.Fatal(err)
		}
		log.Println("✅ Superuser created!!")
		return

	}
	if err != nil {
		log.Fatal(err)
	}

	log.Println("ℹ️ Superuser already exists, skipping seed!!")
}
