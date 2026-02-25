package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite"

	"policyflow/internal/database"
	"policyflow/internal/email"
	"policyflow/internal/handlers"
	authmw "policyflow/internal/middleware"
	"policyflow/internal/seed"
)

//go:embed all:web/out
var webFiles embed.FS

func main() {
	dbPath := getEnv("DB_PATH", "policyflow.db")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-me-in-production")
	port := getEnv("PORT", "8080")

	if os.Getenv("JWT_SECRET") == "" {
		log.Println("WARNING: JWT_SECRET not set — using insecure default (development only)")
	}

	// ── Database ───────────────────────────────────────────────────────────
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1) // SQLite is single-writer

	db := database.New(sqlDB)
	if err := db.Init(); err != nil {
		log.Fatalf("init db: %v", err)
	}
	if err := seed.Run(db); err != nil {
		log.Printf("seed warning: %v", err)
	}

	// ── Services ───────────────────────────────────────────────────────────
	mailer := email.New()
	authMW := authmw.NewAuth(jwtSecret)

	authH := handlers.NewAuth(db, mailer, jwtSecret)
	userH := handlers.NewUser(db, mailer, jwtSecret)
	policyH := handlers.NewPolicy(db)

	// ── Echo ───────────────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	// ── API routes ─────────────────────────────────────────────────────────
	api := e.Group("/api")

	// Public
	api.POST("/magic-link", authH.RequestMagicLink)
	api.GET("/magic-login", authH.MagicLogin)

	// Authenticated (any role)
	authAPI := api.Group("", authMW.Require)
	authAPI.GET("/me", authH.Me)
	authAPI.GET("/policies", policyH.List)
	authAPI.GET("/policies/:id", policyH.Get)
	authAPI.GET("/policies/:id/versions", policyH.Versions)
	authAPI.POST("/policies/:id/acknowledge", policyH.Acknowledge)

	// Admin only
	adminAPI := api.Group("", authMW.Require, authMW.RequireAdmin)
	adminAPI.GET("/users", userH.List)
	adminAPI.POST("/users", userH.Create)
	adminAPI.GET("/admin/stats", policyH.AdminStats)
	adminAPI.POST("/policies", policyH.Create)
	adminAPI.PUT("/policies/:id", policyH.Update)
	adminAPI.POST("/policies/:id/versions", policyH.CreateVersion)

	// ── Frontend ───────────────────────────────────────────────────────────
	if devProxy := os.Getenv("WEB_DEV_PROXY"); devProxy != "" {
		target, err := url.Parse(devProxy)
		if err != nil {
			log.Fatalf("invalid WEB_DEV_PROXY: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		e.Any("/*", echo.WrapHandler(proxy))
		log.Printf("Frontend proxied to %s", devProxy)
	} else {
		subFS, err := fs.Sub(webFiles, "web/out")
		if err != nil {
			log.Fatalf("embed sub FS: %v", err)
		}
		fileServer := http.FileServer(http.FS(subFS))
		e.GET("/*", func(c echo.Context) error {
			path := strings.TrimPrefix(c.Request().URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			// Serve file if it exists, otherwise fall back to index.html (SPA routing)
			if _, err := fs.Stat(subFS, path); err != nil {
				c.Request().URL.Path = "/"
			}
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		})
	}

	log.Printf("PolicyFlow listening on :%s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
