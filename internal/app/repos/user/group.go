package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Group struct {
	ID   uuid.UUID
	Name string
}

type GroupStore interface {
	CreateGroup(ctx context.Context, g Group) (*uuid.UUID, error)
	ReadGroup(ctx context.Context, gid uuid.UUID) (*Group, error)
	DeleteGroup(ctx context.Context, gid uuid.UUID) error
	SearchGroups(ctx context.Context, s string) (chan Group, error)
}

type Groups struct {
	store GroupStore
}

func NewGroups(store GroupStore) *Groups {
	return &Groups{
		store: store,
	}
}

func (gs *Groups) Create(ctx context.Context, g Group) (*Group, error) {
	g.ID = uuid.New()
	id, err := gs.store.CreateGroup(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("create group error: %w", err)
	}
	g.ID = *id
	return &g, nil
}

func (gs *Groups) Read(ctx context.Context, gid uuid.UUID) (*Group, error) {
	g, err := gs.store.ReadGroup(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("read group error: %w", err)
	}
	return g, nil
}

func (gs *Groups) Delete(ctx context.Context, gid uuid.UUID) (*Group, error) {
	g, err := gs.store.ReadGroup(ctx, gid)
	if err != nil {
		return nil, fmt.Errorf("search user error: %w", err)
	}
	return g, gs.store.DeleteGroup(ctx, gid)
}

func (gs *Groups) SearchGroups(ctx context.Context, s string) (chan Group, error) {
	// FIXME: здесь нужно использвоать паттерн Unit of Work
	// бизнес-транзакция
	chin, err := gs.store.SearchGroups(ctx, s)
	if err != nil {
		return nil, err
	}
	chout := make(chan Group, 100)
	go func() {
		defer close(chout)
		for {
			select {
			case <-ctx.Done():
				return
			case u, ok := <-chin:
				if !ok {
					return
				}
				chout <- u
			}
		}
	}()
	return chout, nil
}
