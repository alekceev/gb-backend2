package store

import (
	"gb-backend2/internal/app/repos/user"
	"gb-backend2/internal/db/mem/memstore"
)

type Store struct {
	User      *user.Users
	Group     *user.Groups
	UserGroup *user.UserGroupMapper
}

func NewStore() (*Store, error) {
	var store Store

	s := memstore.NewStore()

	store.User = user.NewUsers(s)
	store.Group = user.NewGroups(s)
	store.UserGroup = user.NewUserGroups(s)

	return &store, nil
}
