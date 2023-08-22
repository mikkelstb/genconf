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

  Example of a conf file:

  <blockname name>
   key value
   key2 value2
   	<subnode name>
     key value
	</subnode>
	<subnode name2>
	 key value
	 # comment
	 # below is a blank line

	</subnode>
  </blockname>
*/

// The following regular expressions are used to parse the configuration file.

var (
	// block_start_pattern matches the start of a block.
	// Example: <blockname>
	block_start_pattern = regexp.MustCompile(`^\s*?<([a-zA-Z0-9]+)>$`)

	// block_end_pattern matches the end of a block.
	// Example: </blockname>
	block_end_pattern = regexp.MustCompile(`^\s*?</([a-zA-Z0-9]+)>$`)

	// attribute_pattern matches an attribute.
	// Example: key value
	attribute_pattern = regexp.MustCompile(`^\s*?(\S+)\s*?(\S*)$`)

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

// ConfNode is an interface that is implemented by all types that can be
// children of a Conf object.

type ConfNode interface {
	String() string
}

// Conf is the main type of the package.
// It represents a block in the configuration file.
type Conf struct {
	name     string
	value    string
	parent   *Conf
	children []ConfNode
}

// Get returns a child node with the given name.
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

// Value returns the value of the attribute with the given name.
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

// Values returns all values of the attribute with the given name.
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
		default:
			panic(fmt.Sprintf("Unknown type: %T", child))
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
// If the file cannot be opened or parsed, it panics.
func ParseFile(filename string) *Conf {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	conf, err := parse(scanner, nil)
	if err != nil {
		panic(err)
	}
	return conf
}

// parse is a recursive function that parses a configuration file.
// It returns a Conf object.
// It reads from the scanner one line at a time.
func parse(scanner *bufio.Scanner, parent *Conf) (*Conf, error) {
	c := &Conf{parent: parent}

	for scanner.Scan() {
		line := scanner.Text()
		if matches := block_start_pattern.FindStringSubmatch(line); matches != nil {
			child, err := parse(scanner, c)
			if err != nil {
				return nil, err
			}
			child.name = matches[1]
			c.children = append(c.children, child)
		} else if matches := block_end_pattern.FindStringSubmatch(line); matches != nil {
			return c, nil
		} else if matches := attribute_pattern.FindStringSubmatch(line); matches != nil {
			c.addAttribute(matches[1], matches[2], "")
		} else if matches := quoted_pattern.FindStringSubmatch(line); matches != nil {
			c.addAttribute(matches[1], matches[2], "\"")
		} else if matches := single_quoted_pattern.FindStringSubmatch(line); matches != nil {
			c.addAttribute(matches[1], matches[2], "'")
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
	return c.parent.indent() + "  "
}
