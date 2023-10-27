package amember

import (
	"fmt"
	"net/url"
)

type User struct {
	Id        int            `json:"user_id"`
	Fob       string         `json:"fob"`
	FobAccess string         `json:"fob_access"`
	Announce  []string       `json:"announce"`
	Nickname  string         `json:"nickname"`
	Nested    usersNestedVal `json:"nested"`
}

type usersNestedVal struct {
	Access []accessNested `json:"access,flow"`
}

type accessNested struct {
	ProductId string `json:"product_id"`
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"expire_date"`
}

func GetUser(id int) (User, error) {
	query := url.Values{}
	query.Set("_nested[]", "access")
	val, err := doRequest[[]User](fmt.Sprintf("/api/users/%d", id), query)
	if err != nil {
		return User{}, err
	}
	return val[0], nil
}

func GetAllUsers() ([]User, error) {
	query := url.Values{}
	query.Set("_nested[]", "access")
	return allPages[User]("/api/users", query)
}
