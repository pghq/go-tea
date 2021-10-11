// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package database provides resources for working with databases.
package database

import (
	"context"
)

// Client represents a client for operating on a database.
type Client interface {
	IsConnected() bool
	Query() Query
	Add() Add
	Remove() Remove
	Transaction(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction.
type Transaction interface {
	Execute(statement Communicator) (uint, error)
	Commit() error
	Rollback() error
}

// Communicator represents a statement encoder
type Communicator interface {
	Statement() (string, []interface{}, error)
}

// Cursor represents an iterator like object for paginated db responses
type Cursor interface {
	Next() bool
	Decode(values ...interface{}) error
	Close()
	Error() error
}

// Query represents a query builder
type Query interface {
	Communicator
	Secondary() Query
	From(collection string) Query
	And(collection string, args ...interface{}) Query
	Filter(filter string, args ...interface{}) Query
	Order(by string) Query
	First(first uint) Query
	After(key string, value interface{}) Query
	Return(key string, args ...interface{}) Query
	Decode(to Query) error
	Execute(ctx context.Context) (Cursor, error)
}

// Add represents a command to add items to the collection
type Add interface {
	Communicator
	To(collection string) Add
	Item(value map[string]interface{}) Add
	Query(query Query) Add
	Decode(to Add) error
	Execute(ctx context.Context) (uint, error)
}

// Remove represents a command to remove items from the collection
type Remove interface {
	Communicator
	From(collection string) Remove
	Filter(filter string, args ...interface{}) Remove
	Order(by string) Remove
	First(first uint) Remove
	After(key string, value interface{}) Remove
	Decode(to Remove) error
	Execute(ctx context.Context) (uint, error)
}
