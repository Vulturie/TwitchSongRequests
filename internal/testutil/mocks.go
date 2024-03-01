package testutil

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/nicklaw5/helix/v2"
	"github.com/vulturie/twitchsongrequests/internal/util"
	"github.com/vulturie/twitchsongrequests/pkg/db"
	"github.com/vulturie/twitchsongrequests/pkg/o11y/metrics"
	"github.com/vulturie/twitchsongrequests/pkg/preferences"
	"github.com/vulturie/twitchsongrequests/pkg/queue"
	"github.com/vulturie/twitchsongrequests/pkg/users"
	"github.com/zmb3/spotify/v2"
)

type DummyPublisher struct {
	Messages          chan string
	ShouldFail        bool
	IsMessageExplicit bool
}

var _ queue.Publisher = (*DummyPublisher)(nil)

func (p DummyPublisher) Publish(client queue.Queuer, url string, pref *preferences.Preference) (spotify.ID, error) {
	if p.ShouldFail {
		return "", errors.New("oops")
	}

	if (pref == nil || !pref.ExplicitSongs) && p.IsMessageExplicit {
		return "", errors.New("not allowed")
	}

	p.Messages <- url
	return "something", nil
}

type MockReadCloser struct{}

func (m MockReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("expected to fail")
}
func (m MockReadCloser) Close() error {
	return nil
}

var _ queue.Queuer = (*MockQueuer)(nil)

type MockQueuer struct {
	ShouldFail   bool
	Messages     []spotify.ID
	Explicit     bool
	GetTrackFunc func(spotify.ID) (*spotify.FullTrack, error)
}

func (m *MockQueuer) QueueSong(ctx context.Context, trackID spotify.ID) error {
	if m.ShouldFail {
		return errors.New("expected to fail")
	}

	m.Messages = append(m.Messages, trackID)
	return nil
}

func (m *MockQueuer) GetTrack(ctx context.Context, id spotify.ID, opts ...spotify.RequestOption) (*spotify.FullTrack, error) {
	return m.GetTrackFunc(id)
}

func (m *MockQueuer) Search(ctx context.Context, query string, t spotify.SearchType, opts ...spotify.RequestOption) (*spotify.SearchResult, error) {
	if m.ShouldFail {
		return nil, errors.New("expected to fail")
	}

	return &spotify.SearchResult{
		Tracks: &spotify.FullTrackPage{
			Tracks: []spotify.FullTrack{
				{
					SimpleTrack: spotify.SimpleTrack{
						ID: "abc123",
					},
				},
			},
		},
	}, nil
}

func DefaultMockQueuerGetTrackFunc(id spotify.ID) (*spotify.FullTrack, error) {
	return &spotify.FullTrack{}, nil
}

// InMemoryUserStore is used for mocking and unit testing
type InMemoryUserStore struct {
	Data map[string]*users.User
}

var _ db.UserStore = (*InMemoryUserStore)(nil)

func (s *InMemoryUserStore) GetUser(id string) (*users.User, error) {
	user, ok := s.Data[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}

	return user, nil
}

func (s *InMemoryUserStore) AddUser(user *users.User) error {
	s.Data[user.TwitchID] = user
	return nil
}

func (s *InMemoryUserStore) UpdateUser(user *users.User) error {
	s.Data[user.TwitchID] = user
	return nil
}

func (s *InMemoryUserStore) DeleteUser(id string) error {
	delete(s.Data, id)
	return nil
}

type DummyCallback struct {
	CallbackExecuted chan bool
	CheckExecuted    chan bool
}

func (c *DummyCallback) Callback(a *util.AuthConfig, u db.UserStore, e *helix.EventSubChannelPointsCustomRewardRedemptionEvent, success bool) error {
	log.Println("received callback request", success)
	c.CallbackExecuted <- success
	return nil
}

func (c *DummyCallback) CheckUser(context.Context, *spotify.Client, *helix.EventSubChannelPointsCustomRewardRedemptionEvent) {
	log.Println("received check user request")
	c.CheckExecuted <- true
}

var _ db.PreferenceStore = (*InMemoryPreferenceStore)(nil)

type InMemoryPreferenceStore struct {
	Data map[string]*preferences.Preference
}

func (s *InMemoryPreferenceStore) GetPreference(id string) (*preferences.Preference, error) {
	p, ok := s.Data[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}

	return p, nil
}

func (s *InMemoryPreferenceStore) AddPreference(p *preferences.Preference) error {
	s.Data[p.TwitchID] = p
	return nil
}

func (s *InMemoryPreferenceStore) UpdatePreference(p *preferences.Preference) error {
	s.Data[p.TwitchID] = p
	return nil
}

func (s *InMemoryPreferenceStore) DeletePreference(id string) error {
	delete(s.Data, id)
	return nil
}

var _ db.MessageCounter = (*InMemoryMessageCounter)(nil)

type InMemoryMessageCounter struct {
	Msgs []*metrics.Message
}

func (c *InMemoryMessageCounter) AddMessage(m *metrics.Message) {
	c.Msgs = append(c.Msgs, m)
}

func (c *InMemoryMessageCounter) TotalMessages() uint64 {
	return uint64(len(c.Msgs))
}

func (c *InMemoryMessageCounter) RunningCount(int) uint64 {
	return uint64(len(c.Msgs))
}

func (c *InMemoryMessageCounter) MessagesForUser(string) []*metrics.Message {
	return c.Msgs
}
