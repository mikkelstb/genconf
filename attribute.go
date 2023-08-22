package genconf

// This file contains the rest of the types used in the package.
// Attribute, BlankLine and Comment are all ConfNodes.

// Type Attribute represents a key-value pair.
// The value can be quoted or unquoted.
// The quote is used when writing the attribute back to a file.
// Only quoted attributes can contain spaces.
type Attribute struct {
	name  string
	value string
	quote string
}

func (a Attribute) String() string {
	if a.quote == "" {
		return a.name + " " + a.value + "\n"
	}
	return a.name + " " + a.quote + a.value + a.quote + "\n"
}

// Type BlankLine represents a blank line.
// It is used when writing the configuration back to a file.
type BlankLine struct{}

func (b BlankLine) String() string {
	return "\n"
}

// Type Comment represents a comment.
// It is used when writing the configuration back to a file.
type Comment string

func (c Comment) String() string {
	return "# " + string(c) + "\n"
}
