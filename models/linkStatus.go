package models

import "time"

type LinkStatus struct {
	Link      string
	Status    string
	Error     error
	DateCheck time.Time
}
