package api

import (
	"crypto/cipher"
	"fmt"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyAuthZHandler struct {
	redirectURL   string
	state         string
	authenticator *spotifyauth.Authenticator
	userStore     db.UserStore
	gcm           cipher.AEAD
}

func NewSpotifyAuthZHandler(url, state string, auth *spotifyauth.Authenticator, userStore db.UserStore, gcm cipher.AEAD) *SpotifyAuthZHandler {
	return &SpotifyAuthZHandler{
		redirectURL:   url,
		state:         state,
		authenticator: auth,
		userStore:     userStore,
		gcm:           gcm,
	}
}

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func (h *SpotifyAuthZHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(constants.TwitchIDCookieKey)
	if err != nil {
		// There is no point in authenticating with Spotify because
		// the Twitch user ID is the primary key for a user
		log.Println("Twitch ID is not available", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "missing Twitch ID cookie")
		return
	}

	if err = cookie.Valid(); err != nil {
		log.Println("cookie expired", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "cookie expired")
		return
	}

	bytes, err := util.DecryptTwitchID(cookie.Value, h.gcm)
	if err != nil {
		log.Println("failed to decrypt cookie value", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to decrypt cookie value")
		return
	}

	userID := string(bytes)
	u, err := h.userStore.GetUser(userID)
	if err != nil {
		log.Println("failed to get user", err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to get user")
		return
	}

	token, err := h.authenticator.Token(r.Context(), "", r)
	if err != nil {
		log.Println("failed to get spotify auth token for user", userID)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "failed to get Spotify auth token")
	}

	// TODO: remove this
	if token != nil {
		log.Println("successfully got Spotify token")
	}

	u.SpotifyAccessToken = token.AccessToken
	u.SpotifyRefreshToken = token.RefreshToken

	err = h.userStore.UpdateUser(u)
	if err != nil {
		log.Println("failed to update user", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "failed to save spotify auth")
	}

	http.Redirect(w, r, h.redirectURL, http.StatusFound)
}
