package render

import (
	//	"bytes"
	//	"log"
	"testing"
)

func TestSplitBookmark(t *testing.T) {
	vals := []string{"inky", "pinky", "blinky", "clyde"}
	r := bookmark(vals)
	expect := []uint32{0, 5, 11, 18, 24}
	for i, v := range expect {
		if r[i] != v {
			t.Fatalf("expected val %v cursor %v, got %v", i, v, r[i])
		}
	}
}

func TestSplitPaginate(t *testing.T) {
	vals := []string{"inky", "pinky", "blinky", "clyde"}
	v := bookmark(vals)
	r, err := paginate(v, 15, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 2 {
		t.Fatalf("expected bookmark len 2, got %v", len(r))
	}
	expect := []uint32{0, 5}
	if len(r[0]) != len(expect) {
		t.Fatalf("expected page 1 len %v, got %v", len(expect), len(r[0]))
	}
	for i, v := range expect {
		if r[0][i] != v {
			t.Fatalf("expected page 1 val %v cursor %v, got %v", i, v, r[0][i])
		}
	}
	expect = []uint32{11, 18, 24}
	if len(r[1]) != len(expect) {
		t.Fatalf("expected page 2 len %v, got %v", len(expect), len(r[1]))
	}
	for i, v := range expect {
		if r[1][i] != v {
			t.Fatalf("expected page 2 val %v cursor %v, got %v", i, v, r[1][i])
		}
	}
}

//func TestSplitMenuPaginate(t *testing.T) {
//	menuCfg := DefaultBrowseConfig()
//	menu := NewMenu().WithBrowseConfig(menuCfg)
//	menu.Put("0", "foo")
//	menu.Put("1", "bar")
//
//	vals := []string{"inky", "pinky", "blinky", "clyde", "tinkywinky", "dipsy", "lala", "pu"}
//	v := bookmark(vals)
////	vv, err := paginate(v, 15, 0, 0)
////	if err != nil {
////		t.Fatal(err)
////	}
//
//	menu = menu.WithPages()
//	menuSizes, err := menu.Sizes()
//	log.Printf("sizes %v", menuSizes)
//	if err != nil {
//		t.Fatal(err)
//	}
//	r, err := paginate(v, 30, menuSizes[1], menuSizes[2])
//	if err != nil {
//		t.Fatal(err)
//	}
//	expect := [][]uint32{
//		[]uint32{0, 5, 11},
//		[]uint32{18},
//		[]uint32{24},
//		[]uint32{35, 41, 46},
//	}
//	if len(r) != len(expect) {
//		t.Fatalf("expected page 1 len %v, got %v", len(expect), len(r))
//	}
//	for i, v := range expect {
//		for j, vv := range v {
//			if r[i][j] != vv {
//				t.Fatalf("value mismatch in [%v][%v]", i, j)
//			}
//		}
//	}
//
//	s := explode(vals, r)
//	expectBytes := append([]byte("inky"), byte(0x00))
//	expectBytes = append(expectBytes, []byte("pinky")...)
//	expectBytes = append(expectBytes, byte(0x00))
//	expectBytes = append(expectBytes, []byte("blinky")...)
//	expectBytes = append(expectBytes, byte(0x0a))
//	expectBytes = append(expectBytes, []byte("clyde")...)
//	expectBytes = append(expectBytes, byte(0x0a))
//	expectBytes = append(expectBytes, []byte("tinkywinky")...)
//	expectBytes = append(expectBytes, byte(0x0a))
//	expectBytes = append(expectBytes, []byte("dipsy")...)
//	expectBytes = append(expectBytes, byte(0x00))
//	expectBytes = append(expectBytes, []byte("lala")...)
//	expectBytes = append(expectBytes, byte(0x00))
//	expectBytes = append(expectBytes, []byte("pu")...)
//
//	if !bytes.Equal([]byte(s), expectBytes) {
//		t.Fatalf("expected:\n\t%s\ngot:\n\t%x\n", expectBytes, s)
//	}
//}
