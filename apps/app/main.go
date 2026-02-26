package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
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
	if err := db.Migrate(); err != nil {
		log.Fatalf("migrate db: %v", err)
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminName := os.Getenv("ADMIN_NAME")
	if err := seed.Run(db, adminEmail, adminName); err != nil {
		log.Printf("seed warning: %v", err)
	}

	// ── Services ───────────────────────────────────────────────────────────
	mailer := email.New()
	authMW := authmw.NewAuth(jwtSecret, db)

	authH := handlers.NewAuth(db, mailer, jwtSecret)
	userH := handlers.NewUser(db, mailer, jwtSecret)
	policyH := handlers.NewPolicy(db)
	deptH := handlers.NewDepartments(db)

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
	authAPI.GET("/departments", deptH.List)
	authAPI.GET("/policies", policyH.List)
	authAPI.GET("/policies/:id", policyH.Get)
	authAPI.GET("/policies/:id/versions", policyH.Versions)
	authAPI.POST("/policies/:id/acknowledge", policyH.Acknowledge)

	// DeptAdmin + SuperAdmin
	deptAdminAPI := api.Group("", authMW.Require, authMW.RequireDeptAdmin)
	deptAdminAPI.POST("/policies", policyH.Create)
	deptAdminAPI.PUT("/policies/:id", policyH.Update)
	deptAdminAPI.POST("/policies/:id/versions", policyH.CreateVersion)
	deptAdminAPI.GET("/users", userH.List)
	deptAdminAPI.POST("/users", userH.Create)
	deptAdminAPI.GET("/admin/stats", policyH.AdminStats)

	// SuperAdmin only
	superAdminAPI := api.Group("", authMW.Require, authMW.RequireSuperAdmin)
	superAdminAPI.POST("/departments", deptH.Create)
	superAdminAPI.PUT("/departments/:id", deptH.Update)
	superAdminAPI.DELETE("/departments/:id", deptH.Delete)
	superAdminAPI.PUT("/users/:id", userH.Update)
	superAdminAPI.DELETE("/users/:id", userH.Delete)

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
		e.GET("/*", func(c echo.Context) error {
			rawPath := strings.TrimPrefix(c.Request().URL.Path, "/")
			if rawPath == "" {
				rawPath = "index.html"
			}
			// Next.js static export with trailingSlash:false generates `page.html`
			// files rather than `page/index.html` directories, so check for both.
			if _, err := fs.Stat(subFS, rawPath); err != nil {
				htmlPath := rawPath + ".html"
				if !strings.Contains(rawPath, ".") {
					if _, err2 := fs.Stat(subFS, htmlPath); err2 == nil {
						rawPath = htmlPath
					} else {
						rawPath = "index.html"
					}
				} else {
					rawPath = "index.html"
				}
			}
			// Serve directly from embed FS to avoid http.FileServer's redirect
			// behaviour of /index.html → / (which causes an infinite 301 loop).
			data, err := fs.ReadFile(subFS, rawPath)
			if err != nil {
				return echo.ErrNotFound
			}
			ct := mime.TypeByExtension(filepath.Ext(rawPath))
			if ct == "" {
				ct = http.DetectContentType(data)
			}
			return c.Blob(http.StatusOK, ct, data)
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
