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

type sqliteTestBinding struct{
	db *sql.DB
}

func newSqliteTestBinding() *sqliteTestBinding {
	return &sqliteTestBinding{nil}
}

func (b *sqliteTestBinding) BeginInstance() gopherbouncedb.UserStorage {
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

func (b *sqliteTestBinding) ClosteInstance(s gopherbouncedb.UserStorage) {
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
