package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Name        string
	Data        string
	Permissions int
}

type UserStore interface {
	CreateUser(ctx context.Context, u User) (*uuid.UUID, error)
	ReadUser(ctx context.Context, uid uuid.UUID) (*User, error)
	DeleteUser(ctx context.Context, uid uuid.UUID) error
	SearchUsers(ctx context.Context, s string) (chan User, error)
}

type Users struct {
	store UserStore
}

func NewUsers(store UserStore) *Users {
	return &Users{
		store: store,
	}
}

func (us *Users) Create(ctx context.Context, u User) (*User, error) {
	u.ID = uuid.New()
	id, err := us.store.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("create user error: %w", err)
	}
	u.ID = *id
	return &u, nil
}

func (us *Users) Read(ctx context.Context, uid uuid.UUID) (*User, error) {
	u, err := us.store.ReadUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("read user error: %w", err)
	}
	return u, nil
}

func (us *Users) Delete(ctx context.Context, uid uuid.UUID) (*User, error) {
	u, err := us.store.ReadUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("search user error: %w", err)
	}
	return u, us.store.DeleteUser(ctx, uid)
}

func (us *Users) SearchUsers(ctx context.Context, s string) (chan User, error) {
	// FIXME: здесь нужно использвоать паттерн Unit of Work
	// бизнес-транзакция
	chin, err := us.store.SearchUsers(ctx, s)
	if err != nil {
		return nil, err
	}
	chout := make(chan User, 100)
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
				u.Permissions = 0755
				chout <- u
			}
		}
	}()
	return chout, nil
}
