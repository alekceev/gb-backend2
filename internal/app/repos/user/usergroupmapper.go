package user

import (
	"context"
	"fmt"
)

type UserGroupsStore interface {
	AddUserToGroup(ctx context.Context, u User, g Group) error
	DeleteUserFromGroup(ctx context.Context, u User, g Group) error
	GetUserGroups(ctx context.Context, u User) (chan Group, error)
	GetGroupUsers(ctx context.Context, g Group) (chan User, error)
}

type UserGroupMapper struct {
	store UserGroupsStore
}

func NewUserGroups(store UserGroupsStore) *UserGroupMapper {
	return &UserGroupMapper{
		store: store,
	}
}

func (ugm *UserGroupMapper) AddUserToGroup(ctx context.Context, u User, g Group) error {
	err := ugm.store.AddUserToGroup(ctx, u, g)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

func (ugm *UserGroupMapper) DeleteUserFromGroup(ctx context.Context, u User, g Group) error {
	err := ugm.store.DeleteUserFromGroup(ctx, u, g)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

func (ugm *UserGroupMapper) GetUserGroups(ctx context.Context, u User) (chan Group, error) {
	ug, err := ugm.store.GetUserGroups(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return ug, nil
}

func (ugm *UserGroupMapper) GetGroupUsers(ctx context.Context, g Group) (chan User, error) {
	gu, err := ugm.store.GetGroupUsers(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return gu, nil
}
