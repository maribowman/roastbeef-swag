package model

import "time"

type ShoppingEntry struct {
	ID     int
	Item   string
	Amount int
	Date   time.Time
}
