package main

import (
	"boiler-plate/internal/core/config"
	"boiler-plate/internal/routes"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/heptiolabs/healthcheck"
)

func main() {
	// docs.SwaggerInfo.Host = config.GetEnvVariable("SWAGGER_HOST", "localhost:8002")
	r := gin.Default()

	db := config.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		panic("Cannot get DB")
	}
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowAllOrigins:  true,
	}))
	//for static file index
	r.Static("/assets", "./dist/assets")
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})
	// swagger
	// r.Use(static.Serve("/", static.LocalFile("./dist", true)))
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	route := routes.NewRoute(r)
	route.Setup()
	startReadinessEndpoints(sqlDB)
	srv := &http.Server{
		Addr:              "0.0.0.0:" + config.GetEnvVariable("PORT", "8000"),
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	config.CloseDB()
}

func startReadinessEndpoints(db *sql.DB) {
	health := healthcheck.NewHandler()
	enableReadinessCheckStr := config.GetEnvVariable("ENABLE_READINESS_CHECKS", "true")
	enableReadinessCheck, _ := strconv.ParseBool(enableReadinessCheckStr)
	if enableReadinessCheck {
		readinessTimeoutStr := config.GetEnvVariable("READINESS_TIMEOUT", "5")
		readinessTimeout, _ := strconv.Atoi(readinessTimeoutStr)
		timeout := time.Duration(readinessTimeout) * time.Second
		health.AddReadinessCheck("database", healthcheck.DatabasePingCheck(db, timeout))
	}

}
