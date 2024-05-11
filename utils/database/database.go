package database

import (
    "gorm.io/gorm"
)

var db *gorm.DB
var DBservice string

func Init(initdb *gorm.DB) {
    if db == nil {
        db = initdb
    }
}

func GetDB() *gorm.DB {
    return db
}
