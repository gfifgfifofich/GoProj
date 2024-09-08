package goproj

type User struct {
	GUID           string   `json:"guid"`
	Email          string   `json:"email"`
	Password       string   `json:"password"`
	RefreshTockens []string `json:"rts"`
}
