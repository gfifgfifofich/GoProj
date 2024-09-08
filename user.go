package goproj

type User struct {
	GUID           string   `json:"guid"`
	Name           string   `json:"name"`
	Password       string   `json:"password"`
	RefreshTockens []string `json:"rts"`
}
