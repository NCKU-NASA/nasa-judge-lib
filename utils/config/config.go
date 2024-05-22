package config

import (
    "os"
    "time"
    "strconv"
)

var Debug bool
var TZ *time.Location
var Secret string
var DBservice string
var DBuser string
var DBpasswd string
var DBhost string
var DBport string
var DBname string
var DBdebug bool

func init() {
    loadenv()
    var err error
    debugstr, exists := os.LookupEnv("DEBUG")
    if !exists {
        Debug = false
    } else {
        Debug, err = strconv.ParseBool(debugstr)
        if err != nil {
            Debug = false
        }
    }
    dbdebugstr, exists := os.LookupEnv("DBDEBUG")
    if !exists {
        DBdebug = true
    } else {
        DBdebug, err = strconv.ParseBool(dbdebugstr)
        if err != nil {
            DBdebug = false
        }
    }
    TZ, err = time.LoadLocation(os.Getenv("TZ"))
    if err != nil {
        panic(err)
    }
    Secret = os.Getenv("SECRET")
    DBservice = os.Getenv("DBSERVICE")
    DBuser = os.Getenv("DBUSER")
    DBpasswd = os.Getenv("DBPASSWD")
    DBhost = os.Getenv("DBHOST")
    DBport = os.Getenv("DBPORT")
    DBname = os.Getenv("DBNAME")
}
