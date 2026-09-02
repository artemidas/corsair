package main

import (
	"embed"
	"log"
	"os"

	"ladon/project"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:dist
var assets embed.FS

func main() {
	dbPath, err := project.DefaultDBPath()
	if err != nil {
		log.Fatal(err)
	}
	projects, err := project.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer project.Close(projects)

	app := application.New(application.Options{
		Name:        "Ladon",
		Description: "Kubernetes security review",
		Services: []application.Service{
			application.NewService(projects),
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
