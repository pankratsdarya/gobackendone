package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
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
	time.Sleep(10 * time.Second)
	log.Printf("Server stopped.")
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func servShutdown(cancel context.CancelFunc) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interrupt
	log.Printf("got signal %q", sig.String())
	messages <- "Server is stopping."
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

func handleConn(conn net.Conn, ctx context.Context) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("Error on closing conn, %v", err)
		}
	}()

	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	input := bufio.NewScanner(conn)
	ch <- "You are " + who
	ch <- "Set nickname? y/n"
	input.Scan()

	if input.Text() == "y" {
		ch <- "Enter your nickname."
		input.Scan()
		who = input.Text()
	}

	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	for input.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			messages <- who + ": " + input.Text()
		}
	}
	leaving <- ch
	messages <- who + " has left"
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
