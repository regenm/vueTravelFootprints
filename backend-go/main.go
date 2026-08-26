package main

import (
	"log"
	"net/http"
	"os"
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

	if err := db.DeleteUserCascade("demo"); err != nil {
		log.Printf("清理测试账号失败: %v", err)
	}
	if err := db.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("管理员账号准备失败: %v", err)
	}
	if err := db.EnsureUser("lime", cfg.LimePassword, "Lime", "lime@travel.local", "user"); err != nil {
		log.Fatalf("账号 lime 准备失败: %v", err)
	}
	if err := db.EnsureUser("eiinxyz", cfg.EiinPassword, "eiinxyz", "eiinxyz@travel.local", "user"); err != nil {
		log.Fatalf("账号 eiinxyz 准备失败: %v", err)
	}

	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	markerHandler := handlers.NewMarkerHandler(db)
	shareHandler := handlers.NewShareHandler(db)
	uploadHandler := handlers.NewUploadHandler(cfg.UploadDir, cfg.PublicURL)
	placeHandler := handlers.NewPlaceHandler(cfg.AmapKey)

	mux := http.NewServeMux()
	protect := middleware.RequireAuth(cfg.JWTSecret)

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"success":true,"data":{"status":"ok"}}`))
	})

	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"message":"未开放注册，请联系管理员开通账号"}`))
	})
	mux.Handle("GET /api/auth/me", protect(http.HandlerFunc(authHandler.Me)))
	mux.Handle("PUT /api/auth/me", protect(http.HandlerFunc(authHandler.UpdateMe)))
	mux.Handle("GET /api/admin/users", protect(http.HandlerFunc(authHandler.ListUsers)))
	mux.Handle("POST /api/admin/users", protect(http.HandlerFunc(authHandler.CreateUser)))

	mux.Handle("GET /api/markers", protect(http.HandlerFunc(markerHandler.GetAll)))
	mux.Handle("GET /api/markers/search", protect(http.HandlerFunc(markerHandler.Search)))
	mux.Handle("GET /api/markers/{id}", protect(http.HandlerFunc(markerHandler.GetByID)))
	mux.Handle("POST /api/markers", protect(http.HandlerFunc(markerHandler.Create)))
	mux.Handle("PUT /api/markers/{id}", protect(http.HandlerFunc(markerHandler.Update)))
	mux.Handle("DELETE /api/markers/{id}", protect(http.HandlerFunc(markerHandler.Delete)))

	mux.Handle("GET /api/places", protect(http.HandlerFunc(placeHandler.Search)))

	mux.Handle("GET /api/shares", protect(http.HandlerFunc(shareHandler.ListMine)))
	mux.Handle("GET /api/shares/inbox", protect(http.HandlerFunc(shareHandler.Inbox)))
	mux.Handle("POST /api/shares", protect(http.HandlerFunc(shareHandler.Create)))
	mux.Handle("PUT /api/shares/{id}", protect(http.HandlerFunc(shareHandler.Update)))
	mux.Handle("DELETE /api/shares/{id}", protect(http.HandlerFunc(shareHandler.Delete)))
	mux.Handle("POST /api/shares/{id}/members", protect(http.HandlerFunc(shareHandler.AddMember)))
	mux.Handle("DELETE /api/shares/{id}/members/{userId}", protect(http.HandlerFunc(shareHandler.RemoveMember)))

	mux.HandleFunc("GET /api/s/{token}", middleware.OptionalAuth(cfg.JWTSecret, shareHandler.PublicGet))

	mux.Handle("POST /api/upload", protect(http.HandlerFunc(uploadHandler.Upload)))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	if cfg.JWTSecret == "travel-footprints-dev-secret-change-me" {
		log.Printf("警告: 正在使用开发用 JWT_SECRET，生产环境请修改")
	}

	if info, err := os.Stat(cfg.StaticDir); err == nil && info.IsDir() {
		mux.HandleFunc("GET /{path...}", handlers.ServeSPA(cfg.StaticDir))
		log.Printf("静态资源: %s", cfg.StaticDir)
	}

	addr := ":" + cfg.Port
	if cfg.Listen != "" {
		addr = cfg.Listen + ":" + cfg.Port
	}
	log.Printf("旅行足迹服务已启动 http://%s", displayListen(cfg.Listen, cfg.Port))
	log.Printf("上传目录: %s", cfg.UploadDir)
	log.Printf("数据库: %s", cfg.DBPath)

	if err := http.ListenAndServe(addr, middleware.CORS(mux)); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func displayListen(listen, port string) string {
	if listen == "" || listen == "0.0.0.0" {
		return "localhost:" + port
	}
	return listen + ":" + port
}
