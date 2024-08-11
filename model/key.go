package model

import (
	"github.com/uptrace/bun"
)

type Key struct {
	bun.BaseModel `bun:"cached"`
	Kid           string `json:"kid"`
	Key           string `json:"key"`
}
