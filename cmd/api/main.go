package main

import (
	"log"
	"net/http"

	"github.com/Dest1on/ProfZoom-backend/internal/config"
	"github.com/Dest1on/ProfZoom-backend/internal/database"
	apphttp "github.com/Dest1on/ProfZoom-backend/internal/http"
)

func main() {
	cfg := config.Load()
	db := database.NewPostgres(cfg.PostgresDSN)
	defer db.Close()

	router := apphttp.NewRouter()

	log.Println("API started on :" + cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, router); err != nil {
		log.Fatal(err)
	}
}
