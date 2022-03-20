package accounts

import (
	"encoding/json"
	"net/http"
	"yatter-backend-go/app/domain/object"
	"yatter-backend-go/app/handler/auth"
	"yatter-backend-go/app/handler/httperror"

	"github.com/go-chi/chi"
)

// Handle request for "POST /v1/accounts/{username}/follow"
func (h *handler) Follow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login := auth.AccountOf(r)
	// TODO: チェックする必要ある?
	if login == nil {
		httperror.InternalServerError(w, nil) //
		return
	}

	targetName := chi.URLParam(r, "username")
	target, err := h.app.Dao.Account().FindByUsername(ctx, targetName)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}
	if target == nil {
		httperror.Error(w, 404)
		return
	}

	relationRepo := h.app.Dao.Relation()
	relation := new(object.RelationWith)
	relation.Following, err = relationRepo.IsFollowing(ctx, login.ID, target.ID)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}
	if !relation.Following {
		if err = relationRepo.Follow(ctx, login.ID, target.ID); err != nil {
			httperror.InternalServerError(w, err)
			return
		}
		relation.Following = true
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
