package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"gb-backend2/internal/api/handler"
	"gb-backend2/internal/api/server"
	"gb-backend2/internal/app/starter"
	"gb-backend2/internal/app/store"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	store, _ := store.NewStore()
	a := starter.NewApp(store)
	h := handler.NewRouter(store)
	srv := server.NewServer(":8000", h, store)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go a.Serve(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
