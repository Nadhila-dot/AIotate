package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	webview "github.com/webview/webview_go"
	"nadhi.dev/sarvar/fun/bootstrap"
	config "nadhi.dev/sarvar/fun/config"
	logg "nadhi.dev/sarvar/fun/logs"
	"nadhi.dev/sarvar/fun/pipeline"
	"nadhi.dev/sarvar/fun/routes"
	"nadhi.dev/sarvar/fun/server"
	sheet "nadhi.dev/sarvar/fun/sheets"
)

const PORT = 317

func init() {
	// Run system checks first
	if err := bootstrap.SystemChecks(); err != nil {
		log.Fatal(err)
	}

	var err error
	var queue_dir string

	queue_dir_val := config.GetConfigValue("SHEET_QUEUE_DIR")
	if queue_dir_val != nil {
		queue_dir = queue_dir_val.(string)
	} else {
		queue_dir = "./storage/queue_data" // Fallback to default
	}

	// Initialize new pipeline system
	pipelineStore, err := pipeline.NewStore("./storage/pipeline")
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to initialize pipeline store: %v", err))
	} else {
		pipelineQueue := pipeline.NewQueue(100, pipelineStore, nil)
		pipelineQueue.Start(context.Background(), 2)
		sheet.GlobalPipelineStore = pipelineStore
		sheet.GlobalPipelineQueue = pipelineQueue
		logg.Success("Pipeline system initialized successfully")
	}

	sheet.GlobalSheetGenerator, err = sheet.NewSheetGenerator(nil, queue_dir, 2)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to initialize GlobalSheetGenerator: %v", err))
		logg.Exit()
	}
	logg.Success("GlobalSheetGenerator initialized successfully")
}

func webserver(port int) {
	log.Fatal(server.Route.Listen(fmt.Sprintf(":%d", port)))
}

func main() {
	// Show banner
	bootstrap.ShowBanner(PORT)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		bootstrap.ShowShutdown()
		os.Exit(0)
	}()

	// Set assets path and register routes
	routes.SetAssetsPath(ExtractedAssetsPath)
	routes.Register()

	// Start web server in background
	go webserver(PORT)

	// Show startup complete message
	bootstrap.ShowStartupComplete()

	// Launch webview
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("AIotate | NADHI.DEV")
	w.SetSize(1024, 768, webview.HintNone)
	w.Navigate(fmt.Sprintf("http://127.0.0.1:%d", PORT))

	w.Run()
}
