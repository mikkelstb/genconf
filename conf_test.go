package genconf_test

import (
	"os"
	"testing"

	"github.com/mikkelstb/genconf"
)

var exampleconf = "./resources/example.conf"
var databaseconf = "./resources/database.conf"

func TestExampleConf(t *testing.T) {

	exconf, err := genconf.ParseFile(exampleconf)
	if err != nil {
		t.Errorf("Error parsing example.conf: %s", err)
	}

	block1 := exconf.Get("block1")

	if block1 == nil {
		t.Errorf("Block1 is nil")
	}

	if block1.Value("key1") != "value1" {
		t.Errorf("Block1 key1 is not value1")
	}

	if block1.Value("key2") != "quoted value" {
		t.Errorf("Block1 key2 is not quoted value")
	}

	if block1.Value("key3") != "single quoted value" {
		t.Errorf("Block1 key3 is not single quoted value")
	}

	block_key4 := block1.Get("key4")

	if block_key4 == nil {
		t.Errorf("Block1 key4 is nil")
	}

	databases := block1.Get("database")
	if databases == nil {
		t.Errorf("databases is nil")
	}

	if len(exconf.Children()) != 2 {
		t.Errorf("exampleconf does not have 2 children")
	}

	if len(databases.Children()) != 2 {
		t.Errorf("databases does not have 2 children")
	}

	maindb := databases.Get("main")
	if maindb == nil {
		t.Errorf("maindb is nil")
	}

	if maindb.Value("user") != "root" {
		t.Errorf("maindb user is not root")
	}

	tasks := block1.Get("tasks").Values("task")
	if len(tasks) != 2 {
		t.Errorf("tasks does not have 2 values")
	}
}

func TestDatabaseConf(t *testing.T) {
	dbconf, err := genconf.ParseFile(databaseconf)
	if err != nil {
		t.Errorf("Error parsing database.conf: %s", err)
	}

	maindbconf := dbconf.Get("main_db")

	if maindbconf == nil {
		t.Errorf("maindb is nil")
	}

	if maindbconf.Value("host") != "corp.example.com" {
		t.Errorf("maindb host is not corp.example.com")
	}

	logger := dbconf.Get("slave").Get("logger").Value("file")

	if logger != "/var/log/db.log" {
		t.Errorf("slave1 logger is not /var/log/db.log")
	}

	keys := dbconf.Get("es").Keys()
	if len(keys) != 2 {
		t.Errorf("es has not 2 Attributes")
	}

	if dbconf.Get("es").Value("host") != "es1.example.com" {
		t.Errorf("es host is not es1.example.com")
	}

	map_of_keys := dbconf.Get("main_db").Map()
	if len(map_of_keys) != 5 {
		t.Errorf("maindb has not 5 Attributes")
	}

	query_1 := "main_db/host"

	if dbconf.GetValueFromPath(query_1) != "corp.example.com" {
		t.Errorf("main_db/host is not corp.example.com")
	}

	query_2 := "slave/logger/file"

	if dbconf.GetValueFromPath(query_2) != "/var/log/db.log" {
		t.Errorf("slave/logger/file is not /var/log/db.log")
	}

	query_3 := "es/user"

	if dbconf.GetValueFromPath(query_3) != "" {
		t.Errorf("es/user is not empty")
	}

	// Test the String() method
	// Read the file and compare it to the string representation of the parsed file

	file, err := os.ReadFile(databaseconf)
	if err != nil {
		t.Errorf("Error reading database.conf: %s", err)
	}

	if string(file) != dbconf.String() {
		t.Errorf("File and string representation are not the same")
	}
}

func TestUnknownFile(t *testing.T) {
	_, err := genconf.ParseFile("unknown.conf")
	if err == nil {
		t.Errorf("Error parsing unknown.conf: %s", err)
	}
}
