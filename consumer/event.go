package consumer

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID   `json:"id"`
	PartyID   uuid.UUID   `json:"party_id"`
	Type      string      `json:"type"`
	Publisher string      `json:"publisher"`
	Payload   interface{} `json:"payload"`
}

func newEvent(bytes []byte) (*Event, error) {
	event := &Event{}
	err := json.Unmarshal(bytes, event)
	return event, err
}
