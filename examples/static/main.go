// Example: Profile data completion menu.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	testdataloader "github.com/peteole/testdata-loader"

	"git.defalsify.org/vise.git/cache"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

var (
	baseDir = testdataloader.GetBasePath()
	scriptDir = path.Join(baseDir, "examples", "static")
	emptyResult = resource.Result{}
)

func out(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: "foo",	
	}, nil
}


func main() {
	var useInternal bool
	root := "root"
	dir := scriptDir
	flag.BoolVar(&useInternal, "i", false, "use internal function for render")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting session at symbol '%s' using resource dir: %s\n", root, dir)

	ctx := context.Background()
	st := state.NewState(0)
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, scriptDir)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(store)
	rs.With(db.DATATYPE_STATICLOAD)

	if useInternal {
		rs.AddLocalFunc("out", out)
	}
	ca := cache.NewCache()
	cfg := engine.Config{
		Root: root,
	}
	en := engine.NewEngine(cfg, rs)
	en = en.WithState(st)
	en = en.WithMemory(ca)

	_, err = en.Exec(ctx, []byte{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
		os.Exit(1)
	}
	_, err = en.Flush(ctx, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "flush error: %v\n", err)
		os.Exit(1)
	}
}
