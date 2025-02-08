package models

import "time"

// Item representa um item dentro de uma lista de compras
type Item struct {
	Checked bool      `firestore:"checked" json:"checked"`
	Created time.Time `firestore:"created" json:"created"`
	Item    string    `firestore:"item" json:"item"`
	User    string    `firestore:"user" json:"user"`
}
