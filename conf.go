package genconf

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

/*
This package is used to parse a configuration file. The configuration file is
of the type Config::General from CPAN.

The config files consist of nested blocks, which can contain attributes of the type key value.
The value can be quoted or unquoted. Only quoted values can contain spaces. Attributes can
be repeated in a block. There are functions to get the value of an attribute, or all values

Comments are allowed on their own line, and blank lines are allowed between blocks as well as key-value lines.

Comments after a key-value or block statements are not supported.

*/

// The following regular expressions are used to parse the configuration file.

var (
	// block_start_pattern matches the start of a block.
	// Example: <block-name>
	block_pattern = regexp.MustCompile(`^\s*?<(\S+?)>$`)

	// blockpattern with a name and value
	// Example: <blockname value>
	block_pattern_with_value = regexp.MustCompile(`^\s*?<(\S+?)\s*?(\S+?)>$`)

	// attribute_comment_pattern matches an attribute with a comment.
	// attribute_comment_pattern = `^\s*?(\S+?)\s*?(["'])?(\S*?)(["'])?((\s*?)(#)(.*))?$`

	// attribute_pattern matches an attribute.
	// Example: key value
	attribute_pattern = regexp.MustCompile(`^\s*?(\S+?)\s*?(\S*)$`)

	// quoted_pattern matches a quoted attribute.
	// Example: key "value"
	quoted_pattern = regexp.MustCompile(`^\s*?(\S+?)\s*?"(.*)"$`)

	// single_quoted_pattern matches a single quoted attribute.
	// Example: key 'value'
	single_quoted_pattern = regexp.MustCompile(`^\s*?(\S+?)\s*?'(.*)'$`)

	// comment_pattern matches a comment.
	// Example: # comment
	comment_pattern = regexp.MustCompile(`^\s*?#(.*)$`)

	// blank_line_pattern matches a blank line.
	blank_line_pattern = regexp.MustCompile(`^\s*$`)
)

var Indent = 4

// ConfNode is an interface that is implemented by all types that can be
// children of a Conf object.

type ConfNode interface {
	String() string
	Name() string
}

// Conf is the main type of the package.
// It represents a block in the configuration file.
type Conf struct {
	name     string
	value    string
	parent   *Conf
	children []ConfNode
}

func (c *Conf) Name() string {
	return c.name
}

// Get returns a child node with the given name.
// If no child node with the given name exists, it returns nil
func (c *Conf) Get(name string) *Conf {
	for _, child := range c.children {
		if conf, ok := child.(*Conf); ok {
			if conf.name == name {
				return conf
			}
		}
	}
	return nil
}

// GetAll returns all child nodes of type Conf with the given name.
func (c *Conf) GetAll(name string) []*Conf {
	var confs []*Conf
	for _, child := range c.children {
		if conf, ok := child.(*Conf); ok {
			if conf.name == name {
				confs = append(confs, conf)
			}
		}
	}
	return confs
}

// Value returns the value of the first attribute with the given name.
func (c *Conf) Value(key string) string {
	for _, child := range c.children {
		if attr, ok := child.(Attribute); ok {
			if attr.name == key {
				return attr.value
			}
		}
	}
	return ""
}

// Values returns all values of the attribute with the given name as a slice of strings.
// If there are no attributes with the given name, an empty slice is returned.
func (c *Conf) Values(key string) []string {
	var values []string
	for _, child := range c.children {
		if attr, ok := child.(Attribute); ok {
			if attr.name == key {
				values = append(values, attr.value)
			}
		}
	}
	return values
}

// Map returns a map of all attributes in the block.
// If there are multiple attributes with the same name, only the last one is
// included in the map.
func (c *Conf) Map() map[string]string {
	m := make(map[string]string)
	for _, child := range c.children {
		if attr, ok := child.(Attribute); ok {
			m[attr.name] = attr.value
		}
	}
	return m
}

// String returns a string representation of the block.
func (c *Conf) String() string {
	sb := strings.Builder{}

	if c.name != "" {
		sb.WriteString("<" + c.name + ">\n")
	}
	for _, child := range c.children {
		switch ch := child.(type) {
		case *Conf:
			sb.WriteString(c.indent() + ch.String())
		case Attribute:
			sb.WriteString(c.indent() + ch.String())
		case Comment:
			sb.WriteString(c.indent() + ch.String())
		case BlankLine:
			sb.WriteString(ch.String())
		}
	}
	if c.name != "" {
		sb.WriteString(c.parent.indent() + "</" + c.name + ">\n")
	}
	return sb.String()
}

func NewConf(name, value string) *Conf {
	return &Conf{
		name:     name,
		value:    value,
		children: []ConfNode{},
		parent:   nil,
	}
}

func (c *Conf) addAttribute(name, value, quote string) {
	c.children = append(c.children, Attribute{name, value, quote})
}

func (c *Conf) addComment(comment string) {
	c.children = append(c.children, Comment(comment))
}

func (c *Conf) addBlankLine() {
	c.children = append(c.children, BlankLine{})
}

// ParseFile parses a configuration file and returns a Conf object.
// If the file cannot be opened or parsed, it returns an error.
func ParseFile(filename string) (*Conf, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	conf, err := parse(scanner, nil)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// check("name") ConfNode
// function checks if the block has a child with the given name
// if so, it returns the child, else returns nil
func (c Conf) check(name string) *Conf {
	for _, child := range c.children {
		if conf, ok := child.(*Conf); ok {
			if conf.name == name {
				return conf
			}
		}
	}
	return nil
}

// parse is a recursive function that parses a configuration file.
// It returns a Conf object.
// It reads from the scanner one line at a time.
func parse(scanner *bufio.Scanner, parent *Conf) (*Conf, error) {
	c := &Conf{parent: parent, children: []ConfNode{}}

	for scanner.Scan() {
		line := scanner.Text()
		if matches := block_pattern.FindStringSubmatch(line); matches != nil {
			//Check if first character is a /, if so, it is a closing block
			if strings.HasPrefix(matches[1], "/") {
				return c, nil
			}
			child, err := parse(scanner, c)
			if err != nil {
				return nil, err
			}
			child.name = matches[1]
			c.children = append(c.children, child)

		} else if matches := block_pattern_with_value.FindStringSubmatch(line); matches != nil {
			// both value and name acts as a block name
			// First check if a block with the same name already exists
			// If so, add the result to the existing block
			// If not, create a new block
			middle := c.check(matches[1])

			// If the block does not exist, create it
			if middle == nil {
				middle = &Conf{parent: c, children: []ConfNode{}}
				middle.name = matches[1]
				c.children = append(c.children, middle)
			}

			// treat the value as a block name, and parse it, adding the result to the middle block
			child, err := parse(scanner, middle)
			if err != nil {
				return nil, err
			}
			child.name = matches[2]
			middle.children = append(middle.children, child)

		} else if matches := quoted_pattern.FindStringSubmatch(line); matches != nil {
			c.addAttribute(matches[1], matches[2], "\"")
		} else if matches := single_quoted_pattern.FindStringSubmatch(line); matches != nil {
			c.addAttribute(matches[1], matches[2], "'")
		} else if matches := attribute_pattern.FindStringSubmatch(line); matches != nil {
			c.addAttribute(matches[1], matches[2], "")
		} else if matches := comment_pattern.FindStringSubmatch(line); matches != nil {
			c.addComment(matches[1])
		} else if matches := blank_line_pattern.FindStringSubmatch(line); matches != nil {
			c.addBlankLine()
		} else {
			return nil, fmt.Errorf("could not parse line: %s", line)
		}
	}
	return c, nil
}

// indent returns a string with the correct indentation for the block.
// It is used when writing the configuration back to a file.
func (c *Conf) indent() string {
	if c.parent == nil {
		return ""
	}
	return c.parent.indent() + strings.Repeat(" ", Indent)
}

// Children() returns a slice of strings with the names of all child blocks.
// Attributes are not included.
// If the block has no children, an empty slice is returned.
func (c Conf) Children() []string {
	children := []string{}
	for _, child := range c.children {
		if conf, ok := child.(*Conf); ok {
			children = append(children, conf.name)
		}
	}
	return children
}

// Keys() returns a slice of strings with the names of all attributes.
// Blocks are not included.
func (c Conf) Keys() []string {
	var keys []string
	for _, child := range c.children {
		if attr, ok := child.(Attribute); ok {
			keys = append(keys, attr.name)
		}
	}
	return keys
}

// GetValueFromPath(path string)
// function returns the value of the attribute with the given xpath like path
// if the path does not exist, it returns an empty string

func (c *Conf) GetValueFromPath(path string) string {
	// Split the path into parts
	parts := strings.Split(path, "/")

	// If the path is empty, return an empty string
	if len(parts) == 0 {
		return ""
	}

	// If the path has only one part, it is the name of the attribute
	if len(parts) == 1 {
		return c.Value(parts[0])
	}

	// If the path has more than one part, the first part is the name of a block
	// and the rest of the parts are the path to the attribute
	block := c.Get(parts[0])
	if block == nil {
		return ""
	}
	return block.GetValueFromPath(strings.Join(parts[1:], "/"))
}
