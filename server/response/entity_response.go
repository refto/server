package response

import "github.com/refto/server/database/model"

type SearchEntity struct {
	Entities      []model.Entity `json:"entities"`
	EntitiesCount int            `json:"entities_count"`
	Topics        []string       `json:"topics"`
}