package links

import "time"

type Link struct {
	ID          int64
	Slug        string
	OriginalURL string
	CreatedAt   time.Time
}
