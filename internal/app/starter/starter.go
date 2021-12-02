package starter

import (
	"context"
	"sync"

	"gb-backend2/internal/app/store"
)

type App struct {
	st *store.Store
}

func NewApp(st *store.Store) *App {
	a := &App{
		st: st,
	}
	return a
}

type APIServer interface {
	Start()
	Stop()
}

func (a *App) Serve(ctx context.Context, wg *sync.WaitGroup, hs APIServer) {
	defer wg.Done()
	hs.Start()
	<-ctx.Done()
	hs.Stop()
}
