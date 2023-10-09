package genconf_test

import (
	"fmt"
	"testing"

	"github.com/mikkelstb/genconf"
)

var exampleconf = "./resources/example.conf"
var databaseconf = "./resources/database.conf"

func TestExampleConf(t *testing.T) {

	exconf := genconf.ParseFile(exampleconf)
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

	if len(databases.Children()) != 2 {
		t.Errorf("databases has not 2 children")
	}

	maindb := databases.Get("main")
	if maindb == nil {
		t.Errorf("maindb is nil")
	}

	if maindb.Value("user") != "root" {
		t.Errorf("maindb user is not root")
	}

	fmt.Println(exconf)

}

func TestDatabaseConf(t *testing.T) {
	dbconf := genconf.ParseFile(databaseconf)

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
}
