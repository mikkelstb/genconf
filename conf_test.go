package genconf_test

import (
	"testing"

	"github.com/mikkelstb/genconf"
)

func TestNewConf(t *testing.T) {

	conf := genconf.ParseFile("example.conf")

	// Get value of key1 in block1
	key1 := conf.Get("block1").Value("key1")
	if key1 != "value1" {
		t.Errorf("Expected value1, got %s", key1)
	}

	// Get all values of key6 in block1
	key1s := conf.Get("block1").Values("key6")
	if len(key1s) != 2 {
		t.Errorf("Expected 2 values, got %d", len(key1s))
	}

	// Get a map of all values in key4
	key4 := conf.Get("block1").Get("key4").Map()["key411"]

	if key4 != "value1" {
		t.Errorf("Expected value1, got %s", key4)
	}

	db_conf := genconf.ParseFile("database.conf").Get("db")
	host := db_conf.Value("host")
	user := db_conf.Value("user")
	pass := db_conf.Value("password")
	db := db_conf.Value("database")

	// check the individual values
	if host != "corp.example.com" {
		t.Errorf("Expected corp.example.com, got %s", host)
	}
	if user != "" {
		t.Errorf("Expected empty_string, got %s", user)
	}
	if pass != "root" {
		t.Errorf("Expected root, got %s", pass)
	}
	if db != "main" {
		t.Errorf("Expected main, got %s", db)
	}

}
