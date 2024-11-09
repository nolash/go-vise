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
)

func TestDumpPg(t *testing.T) {
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

//	store.SetPrefix(db.DATATYPE_USERDATA)
//	err = store.Put(ctx, []byte("bar"), []byte("inky"))
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = store.Put(ctx, []byte("foobar"), []byte("pinky"))
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = store.Put(ctx, []byte("foobarbaz"), []byte("blinky"))
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = store.Put(ctx, []byte("xyzzy"), []byte("clyde"))
//	if err != nil {
//		t.Fatal(err)
//	}

	typMap := pgtype.NewMap()
	k := []byte("foo")
	mockVfd := pgconn.FieldDescription{
		Name: "value",
		DataTypeOID: pgtype.ByteaOID,
		Format: typMap.FormatCodeForOID(pgtype.ByteaOID),
	}
	mockKfd := pgconn.FieldDescription{
		Name: "key",
		DataTypeOID: pgtype.ByteaOID,
		Format: typMap.FormatCodeForOID(pgtype.ByteaOID),
	}
	rows := pgxmock.NewRowsWithColumnDefinition(mockKfd, mockVfd)
	//rows = rows.AddRow([]byte("bar"), []byte("inky"))
	rows = rows.AddRow([]byte("foobar"), []byte("pinky"))
	rows = rows.AddRow([]byte("foobarbaz"), []byte("blinky"))
	//rows = rows.AddRow([]byte("xyzzy"), []byte("clyde"))

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT key, value FROM vvise.kv_vise").WithArgs(k, k[0]).WillReturnRows(rows)
	mock.ExpectCommit()

	o, err := store.Dump(ctx, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	k, _ = o.Next(ctx)
	br, err := base64.StdEncoding.DecodeString(strings.Trim(string(k), "\""))
	if !bytes.Equal(br, []byte("foobar")) {
		t.Fatalf("expected key 'foobar', got %s", br)
	}

	k, _ = o.Next(ctx)
	br, err = base64.StdEncoding.DecodeString(strings.Trim(string(k), "\""))
	if !bytes.Equal(br, []byte("foobarbaz")) {
		t.Fatalf("expected key 'foobarbaz', got %s", br)
	}

	k, _ = o.Next(ctx)
	if k != nil {
		t.Fatalf("expected nil")
	}
}
