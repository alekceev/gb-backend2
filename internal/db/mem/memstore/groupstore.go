package memstore

import (
	"context"
	"database/sql"
	"gb-backend2/internal/app/repos/user"
	"strings"
	"time"

	"github.com/google/uuid"
)

var _ user.GroupStore = &Store{}

func (st *Store) CreateGroup(ctx context.Context, g user.Group) (*uuid.UUID, error) {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	st.g[g.ID] = g
	return &g.ID, nil
}

func (st *Store) ReadGroup(ctx context.Context, uid uuid.UUID) (*user.Group, error) {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	g, ok := st.g[uid]
	if ok {
		return &g, nil
	}
	return nil, sql.ErrNoRows
}

// не возвращает ошибку если не нашли
func (st *Store) DeleteGroup(ctx context.Context, uid uuid.UUID) error {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(st.g, uid)
	delete(st.gu, uid)
	return nil
}

func (st *Store) SearchGroups(ctx context.Context, s string) (chan user.Group, error) {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// FIXME: переделать на дерево остатков

	chout := make(chan user.Group, 100)

	go func() {
		defer close(chout)
		st.Lock()
		defer st.Unlock()
		for _, g := range st.g {
			if strings.Contains(g.Name, s) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					return
				case chout <- g:
				}
			}
		}
	}()

	return chout, nil
}
