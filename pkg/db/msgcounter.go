package db

import (
	"github.com/vulturie/twitchsongrequests/pkg/o11y/metrics"
)

type MessageCounter interface {
	AddMessage(*metrics.Message)
	TotalMessages() uint64
	RunningCount(int) uint64
	MessagesForUser(string) []*metrics.Message
}
