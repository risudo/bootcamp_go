package accounts

import (
	"encoding/json"
	"net/http"
	"yatter-backend-go/app/handler/httperror"

	"github.com/go-chi/chi"
)

// Handler request for "GET /v1/accounts/username/following"
func (h *handler) Following(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := chi.URLParam(r, "username")
	account, err := h.app.Dao.Account().FindByUsername(ctx, username)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}
	if account == nil {
		httperror.Error(w, 404)
		return
	}

	accounts, err := h.app.Dao.Relation().Following(ctx, account.ID)
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}
	if accounts == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}
