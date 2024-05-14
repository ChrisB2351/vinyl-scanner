package main

type Album struct {
	ID     uint64 `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Artist string `json:"artist,omitempty"`
}

func (a *Album) String() string {
	str := `"` + a.Name + `"`
	if a.Artist != "" {
		str += ` by ` + a.Artist
	}
	return str
}

type Log struct {
	Timestamp string `json:"ts,omitempty"`
	Album     *Album `json:"album,omitempty"`
}

type Tag struct {
	ID    string
	Album *Album
}
