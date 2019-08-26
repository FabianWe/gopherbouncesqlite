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
	GetUserS, GetUserByNameS, GetUserByEmailS, InsertUserS, UpdateUserS, DeleteUserS string
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
	return res
}

func (q *SQLiteQueries) Init() []string {
	return q.InitS
}

func (q *SQLiteQueries) GetUser() string {
	return q.GetUserS
}

func (q *SQLiteQueries) GetUserByName() string {
	return q.GetUserByEmailS
}

func (q *SQLiteQueries) GetUserByEmail() string {
	return q.GetUserByEmailS
}

func (q *SQLiteQueries) InsertUser() string {
	return q.InsertUserS
}

func (q *SQLiteQueries) UpdateUser(fields []string) string {
	return q.UpdateUserS
}

func (q *SQLiteQueries) DeleteUser() string {
	return q.DeleteUserS
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

func (b SQLiteBridge) ConvertExistsErr(err error) error {
	if sqliteErr, ok := err.(sqlite.Error); ok && sqliteErr.Code == sqlite.ErrConstraint {
		return gopherbouncedb.NewUserExists(fmt.Sprintf("unique constrained failed: %v", sqliteErr))
	}
	return err
}

func (b SQLiteBridge) ConvertAmbiguousErr(err error) error {
	if sqliteErr, ok := err.(sqlite.Error); ok && sqliteErr.Code == sqlite.ErrConstraint {
		return gopherbouncedb.NewAmbiguousCredentials(fmt.Sprintf("unique constrained failed: %v", sqliteErr))
	}
	return err
}

var (
	DefaultSQLiteUserRowNames = gopherbouncedb.DefaultUserRowNames
)

type SQLiteUserStorage struct {
	*gopherbouncedb.SQLUserStorage
	UpdateFieldsS string
}

func NewSQLiteUserStorage(db *sql.DB, replaceMapping map[string]string) *SQLiteUserStorage {
	queries := NewSQLiteQueries(replaceMapping)
	bridge := NewSQLiteBridge()
	sqlStorage := gopherbouncedb.NewSQLUserStorage(db, queries, bridge)
	res := SQLiteUserStorage{sqlStorage, queries.Replacer.Apply(SQLITE_UPDATE_USER_FIELDS)}
	return &res
}


func (s *SQLiteUserStorage) UpdateUser(id gopherbouncedb.UserID, newCredentials *gopherbouncedb.UserModel, fields []string) error {
	// if internal method not supplied or no fields given: use simple update from sql
	if s.UpdateFieldsS == "" || len(fields) == 0 {
		return s.SQLUserStorage.UpdateUser(id, newCredentials, fields)
	}
	// now perform a more sophisticated update
	updates := make([]string, len(fields))
	args := make([]interface{}, len(fields), len(fields) + 1)
	for i, fieldName := range fields {
		if colName, has := DefaultSQLiteUserRowNames[fieldName]; has {
			updates[i] = colName + "=?"
		} else {
			return fmt.Errorf("Invalid field name \"%s\": Must be a valid field name of the user model", fieldName)
		}
		if arg, argErr := newCredentials.GetFieldByName(fieldName);argErr == nil {
			fieldName = strings.ToLower(fieldName)
			if fieldName == "datejoined" || fieldName == "lastlogin" {
				if t, isTime := arg.(time.Time); isTime {
					arg = s.Bridge.ConvertTime(t)
				} else {
					return fmt.Errorf("DateJoined / LastLogin must be time.Time, got type %v", reflect.TypeOf(arg))
				}
			}
			args[i] = arg
		} else {
			return argErr
		}
	}
	// append id to args
	args = append(args, id)
	// prepare update string
	updateStr := strings.Join(updates, ",")
	// replace updateStr in UpdateFieldS
	stmt := strings.Replace(s.UpdateFieldsS, "$UPDATE_CONTENT$", updateStr, 1)
	// execute statement
	_, err := s.DB.Exec(stmt, args...)
	if err != nil {
		return s.Bridge.ConvertAmbiguousErr(err)
	}
	return nil
}
