package main

import (
	"fmt"
	"log"
	"net/http"
	"travel-footprints/config"
	"travel-footprints/database"
	"travel-footprints/handlers"
	"travel-footprints/middleware"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	if err := db.InsertSeedData(); err != nil {
		log.Printf("种子数据插入失败: %v", err)
	}

	markerHandler := handlers.NewMarkerHandler(db)
	uploadHandler := handlers.NewUploadHandler(cfg.UploadDir, fmt.Sprintf("http://localhost:%s", cfg.Port))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/markers", markerHandler.GetAll)
	mux.HandleFunc("GET /api/markers/search", markerHandler.Search)
	mux.HandleFunc("GET /api/markers/{id}", markerHandler.GetByID)
	mux.HandleFunc("POST /api/markers", markerHandler.Create)
	mux.HandleFunc("PUT /api/markers/{id}", markerHandler.Update)
	mux.HandleFunc("DELETE /api/markers/{id}", markerHandler.Delete)

	mux.HandleFunc("POST /api/upload", uploadHandler.Upload)

	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	handler := middleware.CORS(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 旅行足迹后端服务启动于 http://localhost%s", addr)
	log.Printf("📂 上传目录: %s", cfg.UploadDir)
	log.Printf("🗄️  数据库: %s", cfg.DBPath)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}