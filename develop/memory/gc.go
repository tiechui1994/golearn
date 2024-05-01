package main

type user struct {
	name string
	age  *int
}

func fun1(u *user) {
	u.age = new(int)
}

func fun2(u *user) {
	var b = new(user)
	b.name = "google"
}

func main() {
	u1 := new(user)
	u2 := new(user)

	go fun1(u1)
	go fun2(u2)
}
