package fixtures

const SharedDecision = "one decision written twice"

func alpha() string {
	return SharedDecision
}

func beta() string {
	return SharedDecision
}
