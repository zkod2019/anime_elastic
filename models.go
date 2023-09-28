package main

type Anime struct {
	Id       string `json:"id,omitempty"` // has to be in a string format when u send to elastic
	Title    string `json:"title,omitempty"`
	Author   string `json:"author,omitempty"`
	Season   uint   `json:"season,omitempty"`
	Episodes uint   `json:"episodes,omitempty"`
}
