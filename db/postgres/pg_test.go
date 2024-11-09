package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgconn"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/db/dbtest"
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

	typMap := pgtype.NewMap()

	k := []byte("foo")
	ks := append([]byte{db.DATATYPE_USERDATA}, []byte("foo")...)
	v := []byte("bar")
	resInsert := pgxmock.NewResult("UPDATE", 1)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO vvise.kv_vise").WithArgs(ks, v).WillReturnResult(resInsert)
	mock.ExpectCommit()
	err = store.Put(ctx, k, v)
	if err != nil {
		t.Fatal(err)
	}

	mockVfd := pgconn.FieldDescription{
		Name: "value",
		DataTypeOID: pgtype.ByteaOID,
		Format: typMap.FormatCodeForOID(pgtype.ByteaOID),
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
	br, err := base64.StdEncoding.DecodeString(strings.Trim(string(b), "\""))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(br, v) {
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

	br, err = base64.StdEncoding.DecodeString(strings.Trim(string(b), "\""))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(br, v) {
		t.Fatalf("expected 'plugh', got %x", br)
	}

}
