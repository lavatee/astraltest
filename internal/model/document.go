package model

import (
	"time"
)

type Document struct {
	ID      string    `json:"id" db:"id"`
	Name    string    `json:"name" db:"name"`
	Mime    string    `json:"mime,omitempty" db:"mime"`
	File    bool      `json:"file" db:"is_file"`
	Public  bool      `json:"public" db:"is_public"`
	Created time.Time `json:"created" db:"created_at"`
	Grant   []string  `json:"grant,omitempty"`
}

type DocumentMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime,omitempty"`
	Grant  []string `json:"grant,omitempty"`
}
