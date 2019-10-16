package twitterapi


type User struct {
	Tags []string `toml: "tags"`
}

type Users map[string]*User



