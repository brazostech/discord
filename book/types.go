package book

type Book struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Url            string `json:"url"`
	CurrentChapter int    `json:"current_chapter"`
}
