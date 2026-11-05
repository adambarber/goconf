package fixtures

const Shared = "one decision written twice"

func first() string {
	return Shared
}

func second() string {
	return Shared
}
