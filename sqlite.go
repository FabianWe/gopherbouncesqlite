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

type SQLiteQueries struct {
	InitS                                                                            []string
	GetUserS, GetUserByNameS, GetUserByEmailS, InsertUserS,
		UpdateUserS, DeleteUserS, UpdateFieldsS string
	Replacer *gopherbouncedb.SQLTemplateReplacer
}

func DefaultSQLiteReplacer() *gopherbouncedb.SQLTemplateReplacer {
	return gopherbouncedb.DefaultSQLReplacer()
}

func NewSQLiteQueries(replaceMapping map[string]string) *SQLiteQueries {
	replacer := DefaultSQLiteReplacer()
	if replaceMapping != nil {
		replacer.UpdateDict(replaceMapping)
	}
	res := &SQLiteQueries{}
	res.Replacer = replacer
	// first all init strings
	res.InitS = append(res.InitS, replacer.Apply(SQLITE_USERS_INIT),
		replacer.Apply(SQLITE_USERNAME_INDEX),
		replacer.Apply(SQLITE_USER_EMAIL_INDEX))
	res.GetUserS = replacer.Apply(SQLITE_QUERY_USERID)
	res.GetUserByNameS = replacer.Apply(SQLITE_QUERY_USERNAME)
	res.GetUserByEmailS = replacer.Apply(SQLITE_QUERY_USERMAIL)
	res.InsertUserS = replacer.Apply(SQLITE_INSERT_USER)
	res.UpdateUserS = replacer.Apply(SQLITE_UPDATE_USER)
	res.DeleteUserS = replacer.Apply(SQLITE_DELETE_USER)
	res.UpdateFieldsS = replacer.Apply(SQLITE_UPDATE_USER_FIELDS)
	return res
}

func (q *SQLiteQueries) InitUsers() []string {
	return q.InitS
}

func (q *SQLiteQueries) GetUser() string {
	return q.GetUserS
}

func (q *SQLiteQueries) GetUserByName() string {
	return q.GetUserByNameS
}

func (q *SQLiteQueries) GetUserByEmail() string {
	return q.GetUserByEmailS
}

func (q *SQLiteQueries) InsertUser() string {
	return q.InsertUserS
}

func (q *SQLiteQueries) UpdateUser(fields []string) string {
	if len(fields) == 0 || !q.SupportsUserFields() {
		return q.UpdateUserS
	}
	updates := make([]string, len(fields))
	for i, fieldName := range fields {
		if colName, has := DefaultSQLiteUserRowNames[fieldName]; has {
			updates[i] = colName + "=?"
		} else {
			panic(fmt.Sprintf("invalid field name \"%s\": Must be a valid field name of gopherbouncedb.UserModel", fieldName))
		}
	}
	updateStr := strings.Join(updates, ",")
	stmt := strings.Replace(q.UpdateFieldsS, "$UPDATE_CONTENT$", updateStr, 1)
	return stmt
}

func (q *SQLiteQueries) DeleteUser() string {
	return q.DeleteUserS
}

func (q *SQLiteQueries) SupportsUserFields() bool {
	return q.UpdateFieldsS != ""
}

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

var (
	DefaultSQLiteUserRowNames = gopherbouncedb.DefaultUserRowNames
)

type SQLiteUserStorage struct {
	*gopherbouncedb.SQLUserStorage
}

func NewSQLiteUserStorage(db *sql.DB, replaceMapping map[string]string) *SQLiteUserStorage {
	queries := NewSQLiteQueries(replaceMapping)
	bridge := NewSQLiteBridge()
	sqlStorage := gopherbouncedb.NewSQLUserStorage(db, queries, bridge)
	res := SQLiteUserStorage{sqlStorage}
	return &res
}
