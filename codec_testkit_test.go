package goconf

// Shared types and constants for the codec tests: the config kinds under
// test, the test vocabulary, and the field paths every format walks.

type region string

type host string

type nestedCfg struct {
	Host host `json:"host"`
	Port int  `json:"port"`
}

// caseName is the name of one test case.
type caseName string

// refusalNeed is a fragment an expected refusal must contain.
type refusalNeed string

// none marks a case that must decode without refusal.
const none refusalNeed = ""

// Case names shared across the codec decode tests.
const (
	nameValid   caseName = "valid config decodes"
	nameUnknown caseName = "unknown field is refused naming the field"
)

// Field paths shared across the codec key-walk tests.
const (
	fpRegion     FieldPath = "region"
	fpNested     FieldPath = "nested"
	fpNestedHost FieldPath = "nested.host"
	fpNestedPort FieldPath = "nested.port"
	fpList       FieldPath = "list"
)
