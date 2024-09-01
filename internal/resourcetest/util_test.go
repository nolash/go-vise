package resourcetest

//import (
//	"context"
//	"fmt"
//
//	"git.defalsify.org/vise.git/db"
//	"git.defalsify.org/vise.git/resource"
//)


//type TestSizeResource struct {
//	*DbResource
//}

//func NewTestSizeResource() resource.Resource {
//	rs, err := NewResourceTest()
//	if err != nil {
//		panic(err)
//	}
//
//	rs.AddTemplate(ctx, "small", "one {{.foo}} two {{.bar}} three {{.baz}}")
//	rs.AddTemplate(ctx, "toobug", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus in mattis lorem. Aliquam erat volutpat. Ut vitae metus.")
//	rs.AddTemplate(ctx, "pages", "one {{.foo}} two {{.bar}} three {{.baz}}\n{{.xyzzy}}")
//	rs.AddLocalFunc("foo", get)
//	rs.AddLocalFunc("bar", get)
//	rs.AddLocalFunc("baz", get)
//	rs.AddLocalFunc("xyzzy", getXyzzy)
//	return rs
//}
//
//func get(ctx context.Context, sym string, input []byte) (Result, error) {
//	switch sym {
//	case "foo":
//		return Result{
//			Content: "inky",
//		}, nil
//	case "bar":
//		return Result{
//			Content: "pinky",
//		}, nil
//	case "baz":
//		return Result{
//			Content: "blinky",
//		}, nil
//	}
//	return Result{}, fmt.Errorf("unknown sym: %s", sym)
//}
//
//func getXyzzy(ctx context.Context, sym string, input []byte) (Result, error) {
//	r := "inky pinky\nblinky clyde sue\ntinkywinky dipsy\nlala poo\none two three four five six seven\neight nine ten\neleven twelve"
//	return Result{
//		Content: r,
//	}, nil
//}
