package memstore

import (
	"context"
	"database/sql"
	"gb-backend2/internal/app/repos/user"
	"strings"
	"time"

	"github.com/google/uuid"
)

var _ user.UserStore = &Store{}

func (st *Store) CreateUser(ctx context.Context, u user.User) (*uuid.UUID, error) {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	st.u[u.ID] = u
	return &u.ID, nil
}

func (st *Store) ReadUser(ctx context.Context, uid uuid.UUID) (*user.User, error) {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	u, ok := st.u[uid]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

// не возвращает ошибку если не нашли
func (st *Store) DeleteUser(ctx context.Context, uid uuid.UUID) error {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(st.u, uid)
	delete(st.ug, uid)
	return nil
}

func (st *Store) SearchUsers(ctx context.Context, s string) (chan user.User, error) {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// FIXME: переделать на дерево остатков

	chout := make(chan user.User, 100)

	go func() {
		defer close(chout)
		st.Lock()
		defer st.Unlock()
		for _, u := range st.u {
			if strings.Contains(u.Name, s) {
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					return
				case chout <- u:
				}
			}
		}
	}()

	return chout, nil
}
