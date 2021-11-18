package memstore

import (
	"context"
	"gb-backend2/internal/app/repos/user"
	"time"

	"github.com/google/uuid"
)

var _ user.UserGroupsStore = &Store{}

func (st *Store) AddUserToGroup(ctx context.Context, u user.User, g user.Group) error {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, ok := st.ug[u.ID]; !ok {
		st.ug[u.ID] = make(map[uuid.UUID]struct{})
	}
	if _, ok := st.gu[g.ID]; !ok {
		st.gu[g.ID] = make(map[uuid.UUID]struct{})
	}

	st.ug[u.ID][g.ID] = struct{}{}
	st.gu[g.ID][u.ID] = struct{}{}
	return nil
}

func (st *Store) DeleteUserFromGroup(ctx context.Context, u user.User, g user.Group) error {
	st.Lock()
	defer st.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(st.ug[u.ID], g.ID)
	delete(st.gu[g.ID], u.ID)

	return nil
}

func (st *Store) GetUserGroups(ctx context.Context, u user.User) (chan user.Group, error) {
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
		for i := range st.ug[u.ID] {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				return
			case chout <- st.g[i]:
			}
		}
	}()

	return chout, nil
}

func (st *Store) GetGroupUsers(ctx context.Context, g user.Group) (chan user.User, error) {
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
		for i := range st.gu[g.ID] {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				return
			case chout <- st.u[i]:
			}
		}
	}()

	return chout, nil
}
