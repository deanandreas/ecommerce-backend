package server

import (
	"errors"
	"log/slog"
	"net/http"

	db "github.com/deanandreas/ecommerce-api/internal/database/sqlc"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Server) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	ctx := r.Context()

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("GetUserByID inside GetUserProfile", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to get user profile", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "data fetched successfully", user)
}

func (s *Server) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req struct {
		db.UpdateProfileParams
		Password string `json:"password"`
	}
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.FullName == nil && req.Password == "" && req.Phone == nil && !req.Birth.Valid {
		WriteJSON(w, http.StatusBadRequest, "no filed provieded to update", nil)
		return
	}

	if req.Password != "" {
		if len(req.Password) < 8 {
			WriteJSON(w, http.StatusBadRequest, "invialed password", nil)
			return
		}
		hashPassword, err := HashPassword(req.Password)
		if err != nil {
			slog.Error("failed to hash user password", "error", err)
			WriteJSON(w, http.StatusInternalServerError, "failed to update user profile", nil)
			return
		}
		req.HashPassword = &hashPassword
	}
	req.ID = userID

	ctx := r.Context()
	user, err := s.db.UpdateProfile(ctx, req.UpdateProfileParams)
	if err != nil {
		slog.Error("failed to update user profile", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to update profile data", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "profile updated successfully", user)
}

func (s *Server) AddUserAdress(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req db.InsertAddressParams
	if ok := ReadJSON(w, r, &req); !ok {
		return
	}

	if req.StreetLine1 == "" || req.State == "" || req.City == "" || req.PostalCode == "" || req.Country == "" {
		WriteJSON(w, http.StatusBadRequest, "invialed address data", nil)
		return
	}
	req.UserID = userID

	ctx := r.Context()
	err = s.db.InsertAddress(ctx, req)
	if err != nil {
		slog.Error("failed to insert address", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to add the address", nil)
		return
	}

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("failed to get user by id", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to response the update", nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "address added successfully", user)
}

func (s *Server) UpdateDefaultAddress(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		WriteJSON(w, http.StatusBadRequest, "address id required", nil)
		return
	}

	ctx := r.Context()
	user, err := s.db.UpdateDefaultAddress(ctx, db.UpdateDefaultAddressParams{UserID: userID, ID: id})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid uuid", nil)
				return
			}
		}
		slog.Error("failed to update default address", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to change default address", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "default address updated successfully", user)
}

func (s *Server) DeleteUserAddres(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		WriteJSON(w, http.StatusBadRequest, "address id is reqeired", nil)
		return
	}

	ctx := r.Context()
	row, err := s.db.DeleteUserAddress(ctx, db.DeleteUserAddressParams{UserID: userID, ID: id})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation:
				WriteJSON(w, http.StatusBadRequest, "invalid uuid", nil)
				return
			}
		}
		slog.Error("failed to delete user address", "error", err)
		WriteJSON(w, http.StatusInternalServerError, "failed to delete user address", nil)
		return
	} else if row == 0 {
		WriteJSON(w, http.StatusNotFound, "address not found", nil)
		return
	}

	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("failed to get user by id", "error", err)
		WriteJSON(w, http.StatusBadRequest, "failed to fetch updated user data", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "address deleted successfully", user)
}
