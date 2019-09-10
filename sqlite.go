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
	"fmt"
	"github.com/FabianWe/gopherbouncedb"
	sqlite "github.com/mattn/go-sqlite3"
	"reflect"
	"database/sql"
	"time"
	"strings"
)

var (
	DefaultSQLiteUserRowNames = gopherbouncedb.DefaultUserRowNames
)

// SQLiteBridge implements gopherbouncedb.SQLBridge.
type SQLiteBridge struct{}

func NewSQLiteBridge() SQLiteBridge {
	return SQLiteBridge{}
}

func (b SQLiteBridge) TimeScanType() interface{} {
	var res time.Time
	return &res
}

func (b SQLiteBridge) ConvertTimeScanType(val interface{}) (time.Time, error) {
	switch v := val.(type) {
	case *time.Time:
		return *v, nil
	case time.Time:
		return v, nil
	default:
		var zeroT time.Time
		return zeroT, fmt.Errorf("SQLiteBridge.ConvertTimeScanType: Expected value of *time.Time, got %v",
			reflect.TypeOf(val))
	}
}

func (b SQLiteBridge) ConvertTime(t time.Time) interface{} {
	return t
}

func (b SQLiteBridge) IsDuplicateInsert(err error) bool {
	if sqliteErr, ok := err.(sqlite.Error); ok && sqliteErr.Code == sqlite.ErrConstraint {
		return true
	}
	return false
}

func (b SQLiteBridge) IsDuplicateUpdate(err error) bool {
	if sqliteErr, ok := err.(sqlite.Error); ok && sqliteErr.Code == sqlite.ErrConstraint {
		return true
	}
	return false
}

// SQLiteUserQueries implements gopherbouncedb.UserSQL with support for sqlite3.
type SQLiteUserQueries struct {
	// The slice of init statements.
	// By default contains a create table and two create index (username and email)
	// statements.
	InitS                                                                            []string
	// The default queries.
	GetUserS, GetUserByNameS, GetUserByEmailS, InsertUserS,
		UpdateUserS, DeleteUserS, UpdateFieldsS string
	// The replacer that was used to create the query strings from the strings with
	// meta variables.
	Replacer *gopherbouncedb.SQLTemplateReplacer
	// Used to lookup row names, defaults to DefaultSQLiteUserRowNames.
	RowNames map[string]string
}

// DefaultSQLiteReplacer returns the default sql replacer for sqlite3 (see gopherbouncedb.SQLTemplateReplacer).
// It's the same as gopherbouncedb.DefaultSQLReplacer.
func DefaultSQLiteReplacer() *gopherbouncedb.SQLTemplateReplacer {
	return gopherbouncedb.DefaultSQLReplacer()
}

// NewSQLiteUserQueries returns new queries given the replacement mapping that is used to update
// the default replacer.
//
// That is it uses the default sqlite3 replacer, but updates the fields given in
// replaceMapping to overwrite existing values / insert new ones.
//
// The initialization queries don't contain the creation of an index for the username
//  and email. sqlite automatically creates an index for unique fields in general.
func NewSQLiteUserQueries(replaceMapping map[string]string) *SQLiteUserQueries {
	replacer := DefaultSQLiteReplacer()
	if replaceMapping != nil {
		replacer.UpdateDict(replaceMapping)
	}
	res := &SQLiteUserQueries{}
	res.Replacer = replacer
	// first all init strings
	res.InitS = append(res.InitS, replacer.Apply(SqliteUsersInit))
	res.GetUserS = replacer.Apply(SqliteQueryUserID)
	res.GetUserByNameS = replacer.Apply(SqliteQueryUsername)
	res.GetUserByEmailS = replacer.Apply(SqliteQueryEmail)
	res.InsertUserS = replacer.Apply(SqliteInsertUser)
	res.UpdateUserS = replacer.Apply(SqliteUpdateUser)
	res.DeleteUserS = replacer.Apply(SqliteDeleteUser)
	res.UpdateFieldsS = replacer.Apply(SqliteUpdateUserFields)
	res.RowNames = DefaultSQLiteUserRowNames
	return res
}

func (q *SQLiteUserQueries) InitUsers() []string {
	return q.InitS
}

func (q *SQLiteUserQueries) GetUser() string {
	return q.GetUserS
}

func (q *SQLiteUserQueries) GetUserByName() string {
	return q.GetUserByNameS
}

func (q *SQLiteUserQueries) GetUserByEmail() string {
	return q.GetUserByEmailS
}

func (q *SQLiteUserQueries) InsertUser() string {
	return q.InsertUserS
}

func (q *SQLiteUserQueries) UpdateUser(fields []string) string {
	if len(fields) == 0 || !q.SupportsUserFields() {
		return q.UpdateUserS
	}
	updates := make([]string, len(fields))
	for i, fieldName := range fields {
		if colName, has := q.RowNames[fieldName]; has {
			updates[i] = colName + "=?"
		} else {
			panic(fmt.Sprintf("invalid field name \"%s\": Must be a valid field name of gopherbouncedb.UserModel", fieldName))
		}
	}
	updateStr := strings.Join(updates, ",")
	stmt := strings.Replace(q.UpdateFieldsS, "$UPDATE_CONTENT$", updateStr, 1)
	return stmt
}

func (q *SQLiteUserQueries) DeleteUser() string {
	return q.DeleteUserS
}

func (q *SQLiteUserQueries) SupportsUserFields() bool {
	return q.UpdateFieldsS != ""
}

// SQLiteUserStorage is a user storage based on sqlite3.
type SQLiteUserStorage struct {
	*gopherbouncedb.SQLUserStorage
}

// NewSQLiteUserStorage creates a new sqlite user storage given the database connection
// and the replacement mapping used to create the queries with NewSQLiteUserQueries.
//
// If you want to configure any options please read the gopherbounce wiki.
func NewSQLiteUserStorage(db *sql.DB, replaceMapping map[string]string) *SQLiteUserStorage {
	queries := NewSQLiteUserQueries(replaceMapping)
	bridge := NewSQLiteBridge()
	sqlStorage := gopherbouncedb.NewSQLUserStorage(db, queries, bridge)
	res := SQLiteUserStorage{sqlStorage}
	return &res
}

// SqliteSessionQueries implements gopherbouncedb.SessionSQL with support for sqlite3.
type SqliteSessionQueries struct {
	InitS []string
	InsertSessionS, GetSessionS, DeleteSessionS, CleanUpSessionS, DeleteForUserSessionS string
	Replacer *gopherbouncedb.SQLTemplateReplacer
}

// NewSqliteSessionQueries returns new queries given the replacement mapping that is used to update
// the default replacer.
//
// That is it uses the default sqlite3 replacer, but updates the fields given in
// replaceMapping to overwrite existing values / insert new ones.
func NewSqliteSessionQueries(replaceMapping map[string]string) *SqliteSessionQueries {
	replacer := DefaultSQLiteReplacer()
	if replaceMapping != nil {
		replacer.UpdateDict(replaceMapping)
	}
	res := &SqliteSessionQueries{}
	res.Replacer = replacer
	res.InitS = append(res.InitS, replacer.Apply(SqliteSessionInit))
	res.InsertSessionS = replacer.Apply(SqliteInsertSession)
	res.GetSessionS = replacer.Apply(SqliteGetSession)
	res.DeleteSessionS = replacer.Apply(SqliteDeleteSession)
	res.CleanUpSessionS = replacer.Apply(SqliteCleanUpSession)
	res.DeleteForUserSessionS = replacer.Apply(SqliteDeleteForUser)
	return res
}

func (q *SqliteSessionQueries) InitSessions() []string {
	return q.InitS
}

func (q *SqliteSessionQueries) GetSession() string {
	return q.GetSessionS
}

func (q *SqliteSessionQueries) InsertSession() string {
	return q.InsertSessionS
}

func (q *SqliteSessionQueries) DeleteSession() string {
	return q.DeleteSessionS
}

func (q *SqliteSessionQueries) CleanUpSession() string {
	return q.CleanUpSessionS
}

func (q *SqliteSessionQueries) DeleteForUserSession() string {
	return q.DeleteForUserSessionS
}

// SQLiteSessionStorage is as session storage based on sqlite3.
type SQLiteSessionStorage struct {
	*gopherbouncedb.SQLSessionStorage
}

// NewSQLiteSessionStorage creates a new sqlite session storage given the database connection
// and the replacement mapping used to create the queries with NewSqliteSessionQueries.
//
// If you want to configure any options please read the gopherbounce wiki.
func NewSQLiteSessionStorage(db *sql.DB, replaceMapping map[string]string) *SQLiteSessionStorage {
	queries := NewSqliteSessionQueries(replaceMapping)
	bridge := NewSQLiteBridge()
	sqlStorage := gopherbouncedb.NewSQLSessionStorage(db, queries, bridge)
	return &SQLiteSessionStorage{sqlStorage}
}

// SQLiteStorage combines a user storage and a session storage (both based on sqlite3)
// to implement gopherbouncedb.GoauthStorage.
type SQLiteStorage struct {
	*SQLiteUserStorage
	*SQLiteSessionStorage
}

// NewSQLiteStorage returns a new SQLiteStorage.
func NewSQLiteStorage(db *sql.DB, replaceMapping map[string]string) *SQLiteStorage {
	return &SQLiteStorage{
		NewSQLiteUserStorage(db, replaceMapping),
		NewSQLiteSessionStorage(db, replaceMapping),
	}
}
