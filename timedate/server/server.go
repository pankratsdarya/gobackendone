package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	message = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server started.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go broadcaster()
	go servShutdown(cancel)
	go serverListen(listener, ctx)
	<-ctx.Done()
	time.Sleep(3 * time.Second)
	log.Printf("Server stopped.")

}

func broadcaster() {
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			log.Printf("Error on scaning server message, %v", err)
		}
		message <- msg
	}
}

func servShutdown(cancel context.CancelFunc) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt
	log.Printf("got signal %q", sig.String())
	cancel()
}

func serverListen(listener net.Listener, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			time.Sleep(2 * time.Second)
			return
		default:

			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error on accepting client, %v", err)
				continue
			}
			go handleConn(conn, ctx)
		}
	}
}

func handleConn(c net.Conn, ctx context.Context) {
	defer func() {
		err := c.Close()
		if err != nil {
			log.Printf("Error on closing conn, %v", err)
		}
	}()

	for {

		select {
		case <-ctx.Done():
			_, err := io.WriteString(c, "Server is stopping.")
			if err != nil {
				return
			}
			return
		case msg := <-message:
			_, err := io.WriteString(c, msg)
			if err != nil {
				log.Printf("Error on sending server message, %v", err)
			}
		default:
			_, err := io.WriteString(c, time.Now().Format("15:04:05\n\r"))
			if err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}
