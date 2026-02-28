package main

import (
	"time"

	"gorm.io/gorm"
)

type Album struct {
	gorm.Model
	ID            uint64
	Name          string
	Artist        string
	MusicBrainzID string `gorm:"unique"`
	Tag           string `gorm:"unique"`
}

func (a *Album) String() string {
	str := `"` + a.Name + `"`
	if a.Artist != "" {
		str += ` by ` + a.Artist
	}
	return str
}

type Log struct {
	gorm.Model
	ID      uint64
	Time    time.Time
	AlbumID uint64
	Album   Album
}
