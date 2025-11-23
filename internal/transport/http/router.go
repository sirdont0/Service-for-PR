package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *Handlers) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.HandleFunc("/team/add", h.AddTeam).Methods("POST")
	r.HandleFunc("/team/get", h.GetTeam).Methods("GET")
	r.HandleFunc("/users/setIsActive", h.SetIsActive).Methods("POST")
	r.HandleFunc("/users/getReview", h.GetUserReviews).Methods("GET")
	r.HandleFunc("/pullRequest/create", h.CreatePR).Methods("POST")
	r.HandleFunc("/pullRequest/reassign", h.Reassign).Methods("POST")
	r.HandleFunc("/pullRequest/merge", h.Merge).Methods("POST")
	r.HandleFunc("/statistics/reviewers", h.GetStats).Methods("GET")
	return r
}
