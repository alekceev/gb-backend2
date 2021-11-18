package memstore

import (
	"sync"

	"gb-backend2/internal/app/repos/user"

	"github.com/google/uuid"
)

type Store struct {
	sync.Mutex
	u  map[uuid.UUID]user.User
	g  map[uuid.UUID]user.Group
	ug map[uuid.UUID]map[uuid.UUID]struct{}
	gu map[uuid.UUID]map[uuid.UUID]struct{}
}

func NewStore() *Store {
	return &Store{
		u:  make(map[uuid.UUID]user.User),
		g:  make(map[uuid.UUID]user.Group),
		ug: make(map[uuid.UUID]map[uuid.UUID]struct{}),
		gu: make(map[uuid.UUID]map[uuid.UUID]struct{}),
	}
}
