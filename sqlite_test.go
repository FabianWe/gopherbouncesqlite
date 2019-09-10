// Copyright 2019 Fabian Wenzelmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gopherbouncesqlite

import (
	"database/sql"
	"testing"
	"github.com/FabianWe/gopherbouncedb"
	"github.com/FabianWe/gopherbouncedb/testsuite"
	"os"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFile = "tests.db"
)

func removeDB() {
	if delErr := os.Remove(dbFile); delErr != nil {
		if !os.IsNotExist(delErr) {
			panic(fmt.Sprintf("Can't delete database file: %s", delErr.Error()))
		}
	}
}

type sqliteUserTestBinding struct{
	db *sql.DB
}


func newSqliteTestBinding() *sqliteUserTestBinding {
	return &sqliteUserTestBinding{nil}
}

func (b *sqliteUserTestBinding) BeginInstance() gopherbouncedb.UserStorage {
	// remove existing file
	removeDB()
	// create db
	db, dbErr := sql.Open("sqlite3", dbFile)
	if dbErr != nil {
		panic(fmt.Sprintf("Can't create database: %s", dbErr.Error()))
	}
	b.db = db
	storage := NewSQLiteUserStorage(db, nil)
	return storage
}

func (b *sqliteUserTestBinding) CloseInstance(s gopherbouncedb.UserStorage) {
	if closeErr := b.db.Close(); closeErr != nil {
		panic(fmt.Sprintf("Can't close database: %s", closeErr.Error()))
	}
	removeDB()
}


func TestInit(t *testing.T) {
	testsuite.TestInitSuite(newSqliteTestBinding(), t)
}

func TestInsert(t *testing.T) {
	testsuite.TestInsertSuite(newSqliteTestBinding(), true, t)
}

func TestLookup(t *testing.T) {
	testsuite.TestLookupSuite(newSqliteTestBinding(), true, t)
}

func TestUpdate(t *testing.T) {
	testsuite.TestUpdateUserSuite(newSqliteTestBinding(), true, t)
}

func TestDelete(t *testing.T) {
	testsuite.TestDeleteUserSuite(newSqliteTestBinding(), true, t)
}

// Session testing

type sqliteSessionTestBinding struct {
	db *sql.DB
}

func newSqliteSessionTestBinding() *sqliteSessionTestBinding {
	return &sqliteSessionTestBinding{nil}
}

func (b *sqliteSessionTestBinding) BeginInstance() gopherbouncedb.SessionStorage {
	removeDB()
	// create db
	db, dbErr := sql.Open("sqlite3", dbFile)
	if dbErr != nil {
		panic(fmt.Sprintf("Can't create database: %s", dbErr.Error()))
	}
	b.db = db
	return NewSQLiteSessionStorage(db, nil)
}

func (b *sqliteSessionTestBinding) CloseInstance(storage gopherbouncedb.SessionStorage) {
	if closeErr := b.db.Close(); closeErr != nil {
		panic(fmt.Sprintf("Can't close database: %s", closeErr.Error()))
	}
	removeDB()
}

func TestSessionInit(t *testing.T) {
	testsuite.TestInitSessionSuite(newSqliteSessionTestBinding(), t)
}

func TestSessionInsert(t *testing.T) {
	testsuite.TestSessionInsert(newSqliteSessionTestBinding(), t)
}

func TestSessionGet(t *testing.T) {
	testsuite.TestSessionGet(newSqliteSessionTestBinding(), t)
}

func TestSessionDelete(t *testing.T) {
	testsuite.TestSessionDelete(newSqliteSessionTestBinding(), t)
}

func TestSessionCleanUp(t *testing.T) {
	testsuite.TestSessionCleanUp(newSqliteSessionTestBinding(), t)
}

func TestSessionDeleteForUser(t *testing.T) {
	testsuite.TestSessionDeleteForUser(newSqliteSessionTestBinding(), t)
}
