package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model"
	"github.com/OlayiwolaSherrifSalawu/theTicketing.git/pkg/model/theticket"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	authBodyLimit = 4 << 10
	dbTimeout     = 3 * time.Second
)

type dto struct {
	UserName     string `json:"userName"`
	EmailAddress string `json:"emailAddress"`
	Password     string `json:"password"`
}

// CreateUserHandler registers a user. 201 with the user on success, 422
// on invalid input, 409 when the email or username is taken.
func (a *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input dto
	if err := decodeJSON(w, r, &input, authBodyLimit); err != nil {
		a.clientErrorJSON(w, http.StatusBadRequest, "", "Request body must be valid JSON.")
		return
	}
	if verr := validateSignup(&input); verr != nil {
		a.clientErrorJSON(w, http.StatusUnprocessableEntity, verr.Field, verr.Message)
		return
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		a.serverError(w, err)
		return
	}

	user := &model.User{
		ID:           uuid.New().String(),
		UserName:     input.UserName,
		EmailAddress: input.EmailAddress,
		TimeStamp:    time.Now(),
		HashPassword: hash,
	}
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	if err := a.TheTicket.Insert(ctx, user); err != nil {
		a.storeError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusCreated, user); err != nil {
		a.ErroMessage.Println(err)
	}
}

// LoginUser checks credentials, sets the httpOnly access cookie and also
// returns the token in the body for API clients that use the
// Authorization header instead.
func (a *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	var input dto
	if err := decodeJSON(w, r, &input, authBodyLimit); err != nil {
		a.clientErrorJSON(w, http.StatusBadRequest, "", "Request body must be valid JSON.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()

	user := &model.User{}
	err := a.TheTicket.GetByEmail(ctx, strings.TrimSpace(input.EmailAddress), user)
	if errors.Is(err, theticket.ErrNotFound) {
		a.clientErrorJSON(w, http.StatusUnauthorized, "", "Incorrect email or password.")
		return
	}
	if err != nil {
		a.serverError(w, err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashPassword), []byte(input.Password)) != nil {
		a.clientErrorJSON(w, http.StatusUnauthorized, "", "Incorrect email or password.")
		return
	}

	token, err := a.generateToken(user.ID)
	if err != nil {
		a.serverError(w, err)
		return
	}
	// The cookie must be set before the body is written: headers are
	// frozen once the first byte goes out.
	http.SetCookie(w, a.accessCookie(token, int(tokenLifetime.Seconds())))
	if err := writeJSON(w, http.StatusOK, map[string]string{"token": token}); err != nil {
		a.ErroMessage.Println(err)
	}
}

// LogoutHandler clears the access cookie. JWTs are stateless, so a token
// a client kept elsewhere stays valid until it expires.
func (a *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, a.accessCookie("", -1))
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/")
	}
	w.WriteHeader(http.StatusNoContent)
}

// MeHandler returns the logged-in user.
func (a *Application) MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := extractUserFromContext(r.Context())
	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()
	user, err := a.TheTicket.GetByID(ctx, userID)
	if err != nil {
		a.storeError(w, err)
		return
	}
	if err := writeJSON(w, http.StatusOK, user); err != nil {
		a.ErroMessage.Println(err)
	}
}

func (a *Application) accessCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "access",
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   a.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
