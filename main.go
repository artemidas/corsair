package main

import (
	"embed"
	"log"
	"os"

	"ladon/appdb"
	"ladon/cluster"
	"ladon/project"
	"ladon/rule"
	"ladon/scan"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:dist
var assets embed.FS

func main() {
	dbPath, err := appdb.DefaultPath()
	if err != nil {
		log.Fatal(err)
	}
	db, err := appdb.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer appdb.Close(db)

	session := cluster.NewSession()
	ruleSvc := rule.New(db)
	app := application.New(application.Options{
		Name:        "Ladon",
		Description: "Kubernetes security review",
		Services: []application.Service{
			application.NewService(project.New(db)),
			application.NewService(ruleSvc),
			application.NewService(cluster.New(session)),
			application.NewService(scan.New(db, session, ruleSvc)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Ladon",
		Width:  1800,
		Height: 1000,
		URL:    "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
