package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"yatter-backend-go/app/domain/object"
	"yatter-backend-go/app/handler/auth"
	"yatter-backend-go/app/handler/httperror"

	"github.com/go-chi/chi"
)

// Handler request for "POST /v1/accounts/usernmae/unfollow"
func (h *handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login := auth.AccountOf(r)
	if login == nil {
		httperror.InternalServerError(w, fmt.Errorf("lost account"))
		return
	}

	targetName := chi.URLParam(r, "username")
	target, err := h.app.Dao.Account().FindByUsername(ctx, targetName)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}
	if target == nil {
		httperror.Error(w, http.StatusNotFound)
		return
	}

	relation := &object.RelationShip{
		ID: target.ID,
	}
	relationRepo := h.app.Dao.Relation()
	if err = relationRepo.Unfollow(ctx, login.ID, target.ID); err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	relation.Following, err = relationRepo.IsFollowing(ctx, login.ID, target.ID)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	relation.FollowedBy, err = relationRepo.IsFollowing(ctx, target.ID, login.ID)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(relation); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}
