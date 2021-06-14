package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pankratsdarya/gobackendone/fileserver/fshandlers"
	"github.com/pankratsdarya/gobackendone/fileserver/handlers"
)

func main() {
	handler := &handlers.Handler{}
	uploadHandler := &fshandlers.UploadHandler{
		ServerAddr: "localhost:8080",
		ServeDir:   "./files",
	}
	listHandler := &fshandlers.ListHandler{
		ServeDir: "./files",
	}

	http.Handle("/", handler)
	http.Handle("/upload", uploadHandler)
	http.Handle("/list", listHandler)

	srv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	dirToServe := http.Dir(uploadHandler.ServeDir)
	fs := &http.Server{
		Addr:         ":8080",
		Handler:      http.FileServer(dirToServe),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			fmt.Println("error starting server")
			return
		}
	}()

	go func() {
		err := fs.ListenAndServe()
		if err != nil {
			fmt.Println("error starting server")
			return
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	killsignal := <-interrupt
	switch killsignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Printf("Error while shutting down the server: %v", err)
	} else {
		log.Print("Server stopped.")
	}
	err = fs.Shutdown(ctx)
	if err != nil {
		log.Printf("Error while shutting down the fileserver: %v", err)
	} else {
		log.Print("Fileserver stopped.")
	}

}
