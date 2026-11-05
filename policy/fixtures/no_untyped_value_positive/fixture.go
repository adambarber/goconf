// The positive half: two untyped strings in the var position and an empty
// interface in the parameter position. The rule's other rows are proven
// the same way the count is: one finding per occurrence.
package main

func feed(flags any) {
	var names string

	var more string

	_, _ = names, more

	_, _ = flags, names
}

func main() {
	feed(true)
}
