package main

import (
	"fmt"
	"container/list"
	"strings"
	"time"
)

type User struct {
	Name string
	Ping int
	Flags int
	Client string
	LastTick time.Time
}

type UserList struct {
	ChannelName string
	Users *list.List
}

func NewUser(name string, flags int, ping int, client string) *User {
	user := new(User)
	user.Name = name
	user.Flags = flags
	user.Ping = ping
	user.Client = client
	user.LastTick = time.Now()

	return user
}

func NewUserList(channelName string) *UserList {
	userList := new(UserList)
	userList.ChannelName = channelName
	userList.Users = list.New()

	return userList
}

func (userList *UserList) dump() {
	for u := userList.Users.Front(); u != nil; u = u.Next() {
		user := u.Value.(*User)
		fmt.Println(user)
	}
}

func (userList *UserList) AddUser(name string, flags int, ping int, client string) {
	user := NewUser(name, flags, ping, client)
	userList.Users.PushBack(user)
}

func (userList *UserList) RemoveUser(name string) {
	for u := userList.Users.Front(); u != nil; u = u.Next() {
		user := u.Value.(*User)
		if strings.ToLower(user.Name) == strings.ToLower(name) {
			userList.Users.Remove(u)
			return
		}
	}
}

func (userList *UserList) FindUsers(pattern string) []*User {
	var users []*User

	for u := userList.Users.Front(); u != nil; u = u.Next() {
		user := u.Value.(*User)
		if WildcardCompare(user.Name, pattern) {
			users = append(users, user)
		}
	}
	
	return users
}