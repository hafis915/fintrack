package handler

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/hafis915/fintrack/internal/auth"
	"github.com/hafis915/fintrack/internal/repository"
	"github.com/hafis915/fintrack/pkg/apperr"
	"github.com/hafis915/fintrack/pkg/responses"
)

// localTokenTTL is the lifetime of a Phase 0 local-mint JWT. 24h keeps the
// dev/e2e loop low-friction; v2 (Supabase) will manage real session lifetimes.
const localTokenTTL = 24 * time.Hour

// minPasswordLen is the floor for a local-auth password. Kept modest for a
// single-user Phase 0 app; v2 (Supabase) owns real password policy.
const minPasswordLen = 8

// maxPasswordLen guards bcrypt's 72-byte input limit (longer inputs error).
const maxPasswordLen = 72

// dummyHash equalizes login timing. When the email is unknown we still run a
// bcrypt comparison against this constant so an attacker can't tell "no such
// user" from "wrong password" by response latency (user enumeration).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equalizer-not-a-real-password"), bcrypt.DefaultCost)

// AuthDeps groups the dependencies the local-auth handler needs. Wired in
// internal/server when the routes are mounted (only when AUTH_LOCAL_ENABLED).
type AuthDeps struct {
	Users     repository.UsersRepo
	JWTSecret string
	JWTIssuer string
}

// Auth wires the Phase 0 local-first email register/login routes:
//
//	POST /v1/auth/register
//	POST /v1/auth/login
//
// These are PUBLIC (no JWT) — they exist to MINT the JWT. Passwords are
// bcrypt-hashed at rest; plaintext is never stored or returned. Real Supabase
// Auth replaces this in v2.
type Auth struct {
	d AuthDeps
}

func NewAuth(d AuthDeps) *Auth {
	return &Auth{d: d}
}

// --- request / response shapes ------------------------------------------

type authRegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// --- handlers ------------------------------------------------------------

// Register is POST /v1/auth/register. Creates a user with a real email + name
// and mints a JWT. 409 if the email is already taken.
func (h *Auth) Register(c echo.Context) error {
	var req authRegisterRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}

	name := strings.TrimSpace(req.Name)
	email := normalizeEmail(req.Email)
	if name == "" {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "name is required")
	}
	if !validEmail(email) {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "email is not valid")
	}
	if len(req.Password) < minPasswordLen {
		return responses.Err(c, http.StatusBadRequest, "weak_password", "Password minimal 8 karakter.")
	}
	if len(req.Password) > maxPasswordLen {
		return responses.Err(c, http.StatusBadRequest, "weak_password", "Password maksimal 72 karakter.")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "register_failed", "gagal mendaftar, coba lagi")
	}

	user, err := h.d.Users.Create(c.Request().Context(), email, name, string(hash))
	if err != nil {
		if errors.Is(err, apperr.ErrAlreadyExists) {
			return responses.Err(c, http.StatusConflict, "email_taken", "Email sudah terdaftar — coba masuk saja.")
		}
		return responses.Err(c, http.StatusInternalServerError, "register_failed", "gagal mendaftar, coba lagi")
	}

	token, err := auth.Mint(h.d.JWTSecret, h.d.JWTIssuer, user.ID, localTokenTTL)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "token_mint_failed", "gagal membuat sesi, coba lagi")
	}

	return responses.Created(c, authResponse{
		Token:  token,
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
	})
}

// Login is POST /v1/auth/login. Verifies the email + bcrypt password and mints
// a JWT. Returns a generic 401 for both unknown email and wrong password so the
// endpoint never reveals which emails are registered.
func (h *Auth) Login(c echo.Context) error {
	var req authLoginRequest
	if err := c.Bind(&req); err != nil {
		return responses.Err(c, http.StatusBadRequest, "invalid_json", "could not decode body")
	}

	email := normalizeEmail(req.Email)
	if !validEmail(email) || req.Password == "" {
		return responses.Err(c, http.StatusBadRequest, "invalid_payload", "email dan password wajib diisi")
	}

	user, err := h.d.Users.GetByEmail(c.Request().Context(), email)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			// Run a throwaway compare so unknown-email and wrong-password take
			// the same time — no user enumeration via latency.
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			return responses.Err(c, http.StatusUnauthorized, "invalid_credentials", "Email atau password salah.")
		}
		return responses.Err(c, http.StatusInternalServerError, "login_failed", "gagal masuk, coba lagi")
	}
	// Constant-ish compare; empty hash (legacy/bootstrap rows) can never match.
	if user.PasswordHash == "" ||
		bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return responses.Err(c, http.StatusUnauthorized, "invalid_credentials", "Email atau password salah.")
	}

	token, err := auth.Mint(h.d.JWTSecret, h.d.JWTIssuer, user.ID, localTokenTTL)
	if err != nil {
		return responses.Err(c, http.StatusInternalServerError, "token_mint_failed", "gagal membuat sesi, coba lagi")
	}

	return responses.OK(c, authResponse{
		Token:  token,
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
	})
}

// --- helpers ------------------------------------------------------------

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validEmail(s string) bool {
	if s == "" {
		return false
	}
	addr, err := mail.ParseAddress(s)
	// ParseAddress accepts "Name <addr>" forms; require the bare address.
	return err == nil && addr.Address == s
}
