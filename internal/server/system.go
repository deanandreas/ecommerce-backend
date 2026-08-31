package server

import "net/http"

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/" && r.Method != http.MethodGet {
		WriteJSON(w, http.StatusNotFound, "not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "data fetched successfully", map[string]string{"greerting": "Hello World!"})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.db.Health(r.Context())
	WriteJSON(w, http.StatusOK, "system health fetched successfully", stats)
}
