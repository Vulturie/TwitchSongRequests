package site

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
)

var homePage = template.Must(template.ParseFiles("pkg/site/home.html"))

type HomePageRenderer struct {
	userStore db.UserStore
	twitch    *util.AuthConfig
	spotify   *util.AuthConfig
	siteURL   string
}

type HomePageData struct {
	TwitchAuthURL  string
	SubscribeURL   string
	UnsubscribeURL string
	SpotifyAuthURL string
	Authenticated  bool
}

func NewHomePageRenderer(siteURL string, u db.UserStore, twitch, spotify *util.AuthConfig) *HomePageRenderer {
	return &HomePageRenderer{
		siteURL:   siteURL,
		userStore: u,
		twitch:    twitch,
		spotify:   spotify,
	}
}

func (h *HomePageRenderer) HomePage(w http.ResponseWriter, r *http.Request) {
	data := h.getHomePageData(r)

	if err := homePage.Execute(w, data); err != nil {
		log.Println("error occurred while executing template:", err)
	}
}

func (h *HomePageRenderer) getHomePageData(r *http.Request) *HomePageData {
	d := HomePageData{
		SubscribeURL:   fmt.Sprintf("%s/subscribe", h.siteURL),
		UnsubscribeURL: fmt.Sprintf("%s/revoke", h.siteURL),
		TwitchAuthURL:  fmt.Sprintf("%s/login/twitch", h.siteURL),
		SpotifyAuthURL: util.GenerateAuthURL("accounts.spotify.com", "authorize", h.spotify),
		Authenticated:  false,
	}

	id, err := util.GetUserIDFromRequest(r)
	if err != nil {
		log.Println("failed to get Twitch ID", err)
		return &d
	}

	d.Authenticated = util.IsUserAuthenticated(h.userStore, id)

	return &d
}
