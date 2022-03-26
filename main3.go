package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	group, groupCtx := errgroup.WithContext(ctx)

	linuxSig := make(chan os.Signal, 1)
	signal.Notify(linuxSig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HelloServer)

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	group.Go(func() error {
		return server.ListenAndServe()
	})

	group.Go(func() error {
		<-groupCtx.Done() //阻塞
		fmt.Println("http server stop")
		return server.Shutdown(groupCtx) // 关闭 http server
	})

	group.Go(func() error {
		for {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case <-linuxSig:
				cancelFunc()
			}
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		fmt.Println("group error: ", err)
	}
	fmt.Println("all group done!")
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello, world!\n")
}
