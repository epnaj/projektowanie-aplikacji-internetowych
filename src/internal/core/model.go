package core

import "time"

type ID = uint64

type Link struct {
	Id        ID
	ProjectId ID
	Name      string
	LinkHash  string
	Active    bool
	CreatedAt time.Time
}

type User struct {
	Id           ID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	// MyProject    *Project // single link to users project
	// Links *[]Link
}

type Project struct {
	Id        ID
	OwnerId   ID
	Name      string
	CreatedAt time.Time
}

// type Organisation struct {
// 	Name     string
// 	Projects []*Project
// 	Users    []*User
// }

type Statistic struct {
	Id     ID
	LinkId ID
	Hour   time.Time
	Hits   int64
}
