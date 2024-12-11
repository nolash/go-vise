// Package postgres is a Postgres database backed implementation of the db.Db interface.
//
// The postgres implementation of the vise key-value store uses two data columns of type `BYTEA` for each key and value, aswell as an `updated` field of type `TIMESTAMP` that is set to the current time when an update is made.that is set to the current time when an update is made.
package postgres
