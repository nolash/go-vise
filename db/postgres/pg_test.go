package postgres

import (
	"bytes"
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/db/dbtest"
)

var (
	typMap = pgtype.NewMap()

	mockVfd = pgconn.FieldDescription{
		Name:        "value",
		DataTypeOID: pgtype.ByteaOID,
		Format:      typMap.FormatCodeForOID(pgtype.ByteaOID),
	}
)

func TestCasesPg(t *testing.T) {
	ctx := context.Background()

	t.Skip("implement expects in all cases")

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	store := NewPgDb().WithConnection(mock).WithSchema("vvise")

	err = dbtest.RunTests(t, ctx, store)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutGetPg(t *testing.T) {
	var dbi db.Db
	ses := "xyzzy"

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	store := NewPgDb().WithConnection(mock).WithSchema("vvise")
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(ses)
	ctx := context.Background()

	dbi = store
	_ = dbi

	k := []byte("foo")
	ks := append([]byte{db.DATATYPE_USERDATA}, []byte(ses)...)
	ks = append(ks, []byte(".")...)
	ks = append(ks, k...)
	v := []byte("bar")
	resInsert := pgxmock.NewResult("UPDATE", 1)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectCommit()
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}

	row := pgxmock.NewRowsWithColumnDefinition(mockVfd)
	row = row.AddRow(v)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT value FROM vvise.kv_vise").WithArgs(ks).WillReturnRows(row)
	mock.ExpectCommit()
	b, err := store.Get(ctx, k)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: implement as pgtype map instead, and btw also ask why getting base64 here
	if !bytes.Equal(b, v) {
		t.Fatalf("expected 'bar', got %x", b)
	}

	v = []byte("plugh")
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectCommit()
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectCommit()
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}

	row = pgxmock.NewRowsWithColumnDefinition(mockVfd)
	row = row.AddRow(v)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT value FROM vvise.kv_vise").WithArgs(ks).WillReturnRows(row)
	mock.ExpectCommit()
	b, err = store.Get(ctx, k)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(b, v) {
		t.Fatalf("expected 'plugh', got %x", b)
	}

}

func TestPostgresTxAbort(t *testing.T) {
	var dbi db.Db
	ses := "xyzzy"

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	store := NewPgDb().WithConnection(mock).WithSchema("vvise")
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(ses)
	ctx := context.Background()

	dbi = store
	_ = dbi

	resInsert := pgxmock.NewResult("UPDATE", 1)
	k := []byte("foo")
	ks := append([]byte{db.DATATYPE_USERDATA}, []byte(ses)...)
	ks = append(ks, []byte(".")...)
	ks = append(ks, k...)
	v := []byte("bar")
	//mock.ExpectBegin()
	mock.ExpectBeginTx(defaultTxOptions)
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectRollback()
	err = store.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}
	store.Abort(ctx)
}

func TestPostgresTxCommitOnClose(t *testing.T) {
	var dbi db.Db
	ses := "xyzzy"

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	store := NewPgDb().WithConnection(mock).WithSchema("vvise")
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(ses)
	ctx := context.Background()

	dbi = store
	_ = dbi

	resInsert := pgxmock.NewResult("UPDATE", 1)
	k := []byte("foo")
	ks := append([]byte{db.DATATYPE_USERDATA}, []byte(ses)...)
	ks = append(ks, []byte(".")...)
	ks = append(ks, k...)
	v := []byte("bar")

	ktwo := []byte("blinky")
	kstwo := append([]byte{db.DATATYPE_USERDATA}, []byte(ses)...)
	kstwo = append(kstwo, []byte(".")...)
	kstwo = append(kstwo, ktwo...)
	vtwo := []byte("clyde")

	mock.ExpectBeginTx(defaultTxOptions)
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(kstwo, vtwo).WillReturnResult(resInsert)
	mock.ExpectCommit()

	err = store.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, ktwo, vtwo)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}

	row := pgxmock.NewRowsWithColumnDefinition(mockVfd)
	row = row.AddRow(v)
	mock.ExpectBeginTx(defaultTxOptions)
	mock.ExpectQuery("SELECT value FROM vvise.kv_vise").WithArgs(ks).WillReturnRows(row)
	mock.ExpectCommit()
	row = pgxmock.NewRowsWithColumnDefinition(mockVfd)
	row = row.AddRow(vtwo)
	mock.ExpectBeginTx(defaultTxOptions)
	mock.ExpectQuery("SELECT value FROM vvise.kv_vise").WithArgs(kstwo).WillReturnRows(row)
	mock.ExpectCommit()

	store = NewPgDb().WithConnection(mock).WithSchema("vvise")
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(ses)
	v, err = store.Get(ctx, k)
	if err != nil {
		if !db.IsNotFound(err) {
			t.Fatalf("get key one: %x", k)
		}
	}
	v, err = store.Get(ctx, ktwo)
	if err != nil {
		if !db.IsNotFound(err) {
			t.Fatalf("get key two: %x", ktwo)
		}
	}
}

func TestPostgresTxStartStop(t *testing.T) {
	var dbi db.Db
	ses := "xyzzy"

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	store := NewPgDb().WithConnection(mock).WithSchema("vvise")
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(ses)
	ctx := context.Background()

	dbi = store
	_ = dbi

	resInsert := pgxmock.NewResult("UPDATE", 1)
	k := []byte("inky")
	ks := append([]byte{db.DATATYPE_USERDATA}, []byte(ses)...)
	ks = append(ks, []byte(".")...)
	ks = append(ks, k...)
	v := []byte("pinky")

	ktwo := []byte("blinky")
	kstwo := append([]byte{db.DATATYPE_USERDATA}, []byte(ses)...)
	kstwo = append(kstwo, []byte(".")...)
	kstwo = append(kstwo, ktwo...)
	vtwo := []byte("clyde")
	mock.ExpectBeginTx(defaultTxOptions)
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(kstwo, vtwo).WillReturnResult(resInsert)
	mock.ExpectCommit()

	row := pgxmock.NewRowsWithColumnDefinition(mockVfd)
	row = row.AddRow(v)
	mock.ExpectBeginTx(defaultTxOptions)
	mock.ExpectQuery("SELECT value FROM vvise.kv_vise").WithArgs(ks).WillReturnRows(row)
	row = pgxmock.NewRowsWithColumnDefinition(mockVfd)
	row = row.AddRow(vtwo)
	mock.ExpectQuery("SELECT value FROM vvise.kv_vise").WithArgs(kstwo).WillReturnRows(row)
	mock.ExpectCommit()

	err = store.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(ctx, ktwo, vtwo)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Stop(ctx)
	if err != nil {
		t.Fatal(err)
	}

	v, err = store.Get(ctx, k)
	if err != nil {
		t.Fatal(err)
	}
	v, err = store.Get(ctx, ktwo)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
