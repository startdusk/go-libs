package types

import (
	"fmt"
)

type User struct {
	Name string

	// 因为是在同一个包, 所以未导出字段能访问到
	// 但在不同的包内就访问不到了
	age int
}

func NewUser(name string, age int) User {
	return User{
		Name: name,
		age:  age,
	}
}

func NewUserPtr(name string, age int) *User {
	return &User{
		Name: name,
		age:  age,
	}
}

func (u User) GetAge() int {
	return u.age
}

func (u *User) ChangeName(name string) {
	u.Name = name
}

func (u User) private() {
	fmt.Println("private")
}
