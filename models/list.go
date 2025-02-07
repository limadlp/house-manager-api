package models

import "time"

type ShoppingList struct {
	ID      string    `json:"id,omitempty"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Items   []Item    `json:"items"`
}
