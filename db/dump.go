package db

import (
	"context"
)

type DumperFunc func(ctx context.Context) ([]byte, []byte)

type Dumper struct {
	fn DumperFunc 
	k []byte
	v []byte
	nexted bool
}

func NewDumper(fn DumperFunc) *Dumper {
	return &Dumper{
		fn: fn,
	}
}

func(d *Dumper) WithFirst(k []byte, v []byte) *Dumper {
	if d.nexted {
		panic("already started")
	}
	d.k = k
	d.v = v
	d.nexted = true
	return d
}

func(d *Dumper) Next(ctx context.Context) ([]byte, []byte) {
	d.nexted = true
	k := d.k
	v := d.v
	if k == nil {
		return nil, nil
	}
	d.k, d.v = d.fn(ctx)
	logg.TraceCtxf(ctx, "next value is", "k", d.k, "v", d.v)
	return k, v
}
