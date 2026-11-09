// Command server runs the Nuolydis HTTP API and front end.
package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/api"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/infra/clock"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/infra/mysql"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/infra/random"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
	"github.com/SleepiCaffeine/VGTU-PSI/web"
)

const readHeaderTimeout = 5 * time.Second

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	wallClock, err := clock.New()
	if err != nil {
		return err
	}
	db, err := mysql.Open(envOr("MYSQL_DSN", mysql.DefaultDSN))
	if err != nil {
		return err
	}
	err = mysql.Migrate(db)
	if err != nil {
		return err
	}
	err = mysql.Seed(db, wallClock.Now())
	if err != nil {
		return err
	}
	static, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return fmt.Errorf("open front end: %w", err)
	}

	deps := service.Dependencies{
		UnitOfWork: mysql.NewUnitOfWork(db),
		Clock:      wallClock,
		Engine:     engine.Default(),
		Games:      game.Default(),
		Random:     random.Source{},
	}
	router := api.NewRouter(api.Services{
		Tenants:  service.NewTenantService(deps),
		Plans:    service.NewPlanService(deps),
		Sessions: service.NewSessionService(deps),
		Betting:  service.NewBettingService(deps),
		Alerts:   service.NewAlertService(deps),
	}, api.Options{DemoMode: envOr("DEMO_MODE", "true") == "true", Static: static})

	address := envOr("HTTP_ADDR", ":8080")
	log.Printf("Nuolydis listening on %s", address)
	server := &http.Server{Addr: address, Handler: router, ReadHeaderTimeout: readHeaderTimeout}
	return server.ListenAndServe()
}

func envOr(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
