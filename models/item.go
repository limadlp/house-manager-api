package models

import "time"

// Item representa um item dentro de uma lista de compras
type Item struct {
	Checked bool      `json:"checked"`
	Created time.Time `json:"created"`
	Name    string    `json:"item"` // Alterado de "Item" para "Name"
	User    string    `json:"user"`
}
