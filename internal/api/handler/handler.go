package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gb-backend2/internal/app/repos/user"
	"gb-backend2/internal/app/store"

	"github.com/google/uuid"
)

type Router struct {
	*http.ServeMux
	store *store.Store
}

func NewRouter(store *store.Store) *Router {
	r := &Router{
		ServeMux: http.NewServeMux(),
		store:    store,
	}
	r.Handle("/user/create",
		r.AuthMiddleware(
			http.HandlerFunc(r.CreateUser),
		),
	)
	r.Handle("/user/read", r.AuthMiddleware(http.HandlerFunc(r.ReadUser)))
	r.Handle("/user/delete", r.AuthMiddleware(http.HandlerFunc(r.DeleteUser)))
	r.Handle("/user/search", r.AuthMiddleware(http.HandlerFunc(r.SearchUser)))
	r.Handle("/user/get_groups", r.AuthMiddleware(http.HandlerFunc(r.GetGroups)))
	r.Handle("/user/add_group", r.AuthMiddleware(http.HandlerFunc(r.AddUserToGroup)))
	r.Handle("/user/delete_group", r.AuthMiddleware(http.HandlerFunc(r.DeleteUserFromGroup)))

	r.Handle("/group/create",
		r.AuthMiddleware(
			http.HandlerFunc(r.CreateGroup),
		),
	)
	r.Handle("/group/read", r.AuthMiddleware(http.HandlerFunc(r.ReadGroup)))
	r.Handle("/group/delete", r.AuthMiddleware(http.HandlerFunc(r.DeleteGroup)))
	r.Handle("/group/search", r.AuthMiddleware(http.HandlerFunc(r.SearchGroup)))
	r.Handle("/group/add_user", r.AuthMiddleware(http.HandlerFunc(r.AddUserToGroup)))
	r.Handle("/group/delete_user", r.AuthMiddleware(http.HandlerFunc(r.DeleteUserFromGroup)))

	return r
}

type User struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Data       string    `json:"data"`
	Permission int       `json:"perms"`
}

type Group struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (rt *Router) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if u, p, ok := r.BasicAuth(); !ok || !(u == "admin" && p == "admin") {
				http.Error(w, "unautorized", http.StatusUnauthorized)
				return
			}
			// r = r.WithContext(context.WithValue(r.Context(), 1, 0))
			next.ServeHTTP(w, r)
		},
	)
}

func (rt *Router) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	u := User{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	bu := user.User{
		Name: u.Name,
		Data: u.Data,
	}

	nbu, err := rt.store.User.Create(r.Context(), bu)
	if err != nil {
		http.Error(w, "error when creating", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(
		User{
			ID:         nbu.ID,
			Name:       nbu.Name,
			Data:       nbu.Data,
			Permission: nbu.Permissions,
		},
	)
}

// read?uid=...
func (rt *Router) ReadUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.store.User.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(
		User{
			ID:         nbu.ID,
			Name:       nbu.Name,
			Data:       nbu.Data,
			Permission: nbu.Permissions,
		},
	)
}

func (rt *Router) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.store.User.Delete(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(
		User{
			ID:         nbu.ID,
			Name:       nbu.Name,
			Data:       nbu.Data,
			Permission: nbu.Permissions,
		},
	)
}

// /search?q=...
func (rt *Router) SearchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ch, err := rt.store.User.SearchUsers(r.Context(), q)
	if err != nil {
		http.Error(w, "error when reading", http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)

	first := true
	fmt.Fprintf(w, "[")
	defer fmt.Fprintln(w, "]")

	for {
		select {
		case <-r.Context().Done():
			return
		case u, ok := <-ch:
			if !ok {
				return
			}
			if first {
				first = false
			} else {
				fmt.Fprintf(w, ",")
			}
			_ = enc.Encode(
				User{
					ID:         u.ID,
					Name:       u.Name,
					Data:       u.Data,
					Permission: u.Permissions,
				},
			)
			w.(http.Flusher).Flush()
		}
	}
}

func (rt *Router) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	u := User{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	gu := user.Group{
		Name: u.Name,
	}

	ngu, err := rt.store.Group.Create(r.Context(), gu)
	if err != nil {
		http.Error(w, "error when creating", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(
		Group{
			ID:   ngu.ID,
			Name: ngu.Name,
		},
	)
}

// read?uid=...
func (rt *Router) ReadGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ngu, err := rt.store.Group.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(
		Group{
			ID:   ngu.ID,
			Name: ngu.Name,
		},
	)
}

func (rt *Router) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.store.Group.Delete(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(
		Group{
			ID:   nbu.ID,
			Name: nbu.Name,
		},
	)
}

// /search?q=...
func (rt *Router) SearchGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ch, err := rt.store.Group.SearchGroups(r.Context(), q)
	if err != nil {
		http.Error(w, "error when reading", http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)

	first := true
	fmt.Fprintf(w, "[")
	defer fmt.Fprintln(w, "]")

	for {
		select {
		case <-r.Context().Done():
			return
		case u, ok := <-ch:
			if !ok {
				return
			}
			if first {
				first = false
			} else {
				fmt.Fprintf(w, ",")
			}
			_ = enc.Encode(
				Group{
					ID:   u.ID,
					Name: u.Name,
				},
			)
			w.(http.Flusher).Flush()
		}
	}
}

func (rt *Router) GetGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user, err := rt.store.User.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	ch, err := rt.store.UserGroup.GetUserGroups(r.Context(), *user)
	if err != nil {
		http.Error(w, "error when reading", http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)

	first := true
	fmt.Fprintf(w, "[")
	defer fmt.Fprintln(w, "]")

	for {
		select {
		case <-r.Context().Done():
			return
		case u, ok := <-ch:
			if !ok {
				return
			}
			if first {
				first = false
			} else {
				fmt.Fprintf(w, ",")
			}
			_ = enc.Encode(
				Group{
					ID:   u.ID,
					Name: u.Name,
				},
			)
			w.(http.Flusher).Flush()
		}
	}
}

func (rt *Router) AddUserToGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sgid := r.URL.Query().Get("gid")
	if sgid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	gid, err := uuid.Parse(sgid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (gid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user, err := rt.store.User.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	group, err := rt.store.Group.Read(r.Context(), gid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	err = rt.store.UserGroup.AddUserToGroup(r.Context(), *user, *group)
	if err != nil {
		http.Error(w, "error add group", http.StatusInternalServerError)
	}

	fmt.Fprintln(w, `{"status":"ok"}`)
}

func (rt *Router) DeleteUserFromGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sgid := r.URL.Query().Get("gid")
	if sgid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	gid, err := uuid.Parse(sgid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (gid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user, err := rt.store.User.Read(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	group, err := rt.store.Group.Read(r.Context(), gid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	err = rt.store.UserGroup.DeleteUserFromGroup(r.Context(), *user, *group)
	if err != nil {
		http.Error(w, "error add group", http.StatusInternalServerError)
	}

	fmt.Fprintln(w, `{"status":"ok"}`)
}
