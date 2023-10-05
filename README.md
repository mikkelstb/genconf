Genconf is a library for reading and writing configuration files.

The config files are used in a variety of projects. They contain blocks of key-value pairs, where the keys are strings and the values are intepreted as strings. Blocks can be nested, and the same key can appear multiple times in a block.

Example config file:

<maindb>
    host localhost
    port 5432
    user postgres
    password postgres
    dbname postgres
    <log>
        level debug
        file /var/log/maindb.log
    </log>
</maindb>

## Usage:
The function ReadFile reads a config file and returns a Config object:

 conf := genconf.ParseFile("config.conf")

## Accesing blocks
The method Get returns a Config object for the block with the given name:

 maindb := conf.Get("maindb")

Currently the function panics if the file cannot be read or parsed.

## Accessing values
Use the Value() function to get the value of a key:

 host := conf.Value("host")
 user := conf.Value("user")
 pswd := conf.Value("password")

