package render

import (
	"fmt"
	"strings"
)

func bookmark(values []string) []uint32 {
	var c int
	var bookmarks []uint32 = []uint32{0}

	for _, v := range values {
		c += len(v) + 1
		bookmarks = append(bookmarks, uint32(c))
	}
	return bookmarks
}
//
//func paginate(bookmarks []uint32, capacity uint32) ([][]uint32, error) {
//	if len(bookmarks) == 0 {
//		return nil, fmt.Errorf("empty bookmark array")
//	}
//	var c uint32
//	var pages [][]uint32
//	var prev uint32
//	
//	pages = append(pages, []uint32{})
//	currentPage := 0
//	lookAhead := bookmarks[1:]
//
//	for i, v := range lookAhead {
//		lc := v - c
//		if lc > capacity {
//			c = prev
//			if lc - c > capacity {
//				return nil, fmt.Errorf("value at %v alone exceeds capacity", i)
//			}
//			currentPage += 1
//			pages = append(pages, []uint32{})
//		}
//		pages[currentPage] = append(pages[currentPage], bookmarks[i])
//		prev = v
//	}
//
//	pages[currentPage] = append(pages[currentPage], bookmarks[len(bookmarks)-1])
//	return pages, nil
//}

func isLast(cursor uint32, end uint32, capacity uint32) bool {
	l := end - cursor
	remaining := capacity
	return l <= remaining
}

func paginate(bookmarks []uint32, capacity uint32, nextSize uint32, prevSize uint32) ([][]uint32, error) {
	if len(bookmarks) == 0 {
		return nil, fmt.Errorf("empty page array")
	}

	var pages [][]uint32
	var c uint32
	lastIndex := len(bookmarks) - 1
	last := bookmarks[lastIndex]
	var haveMore bool

	if isLast(0, last, capacity) {
		pages = append(pages, bookmarks)
		return pages, nil
	}

	lookAhead := bookmarks[1:]
	pages = append(pages, []uint32{})
	var i int

	haveMore = true
	for haveMore {
		remaining := int(capacity)
		if i > 0 {
			remaining -= int(prevSize)
		}
		if remaining < 0 {
			return nil, fmt.Errorf("underrun in item %v:%v (%v) index %v prevsize %v remain %v cap %v", bookmarks[i], lookAhead[i], lookAhead[i] - bookmarks[i], i, prevSize, remaining, capacity)
		}
		if isLast(c, last, uint32(remaining)) {
			haveMore = false
		} else {
			remaining -= int(nextSize)
		}
		if remaining < 0 {
			return nil, fmt.Errorf("underrun in item %v:%v (%v) index %v prevsize %v nextsize %v remain %v cap %v", bookmarks[i], lookAhead[i], lookAhead[i] - bookmarks[i], i, prevSize, nextSize, remaining, capacity)
		}

		var z int
		currentPage := len(pages) - 1
		for i < lastIndex {
			logg.Tracef("have render", "bookmark", bookmarks[i], "lookahead", lookAhead[i], "diff", lookAhead[i] - bookmarks[i], "index", i, "prevsize", prevSize, "nextsize", nextSize, "remain", remaining, "capacity", capacity)

			v := lookAhead[i]
			delta := int((v - c) + 1)
			if z == 0 {
				if delta > remaining {
					return nil, fmt.Errorf("single value at %v exceeds capacity", i)
				}
			}
			z += delta
			if z > remaining {
				break
			}
			pages[currentPage] = append(pages[currentPage], bookmarks[i])
			c = v
			i += 1
		}
		logg.Tracef("render more", "have", haveMore, "remain", remaining, "c", c, "last", last, "pages", pages)

		if haveMore {
			pages = append(pages, []uint32{})
		}
	}

	l := len(pages)-1
	pages[l] = append(pages[l], last)
	return pages, nil
}

func explode(values []string, pages [][]uint32) string {
	s := strings.Join(values, "")
	s += "\n"
	sb := strings.Builder{}

	var start uint32
	var end uint32
	var lastPage int
	var z uint32
	for i, page := range pages {
		for _, c := range page {
			if c == 0 {
				continue
			}
			z += 1
			if i != lastPage {
				sb.WriteRune('\n')
			} else if c > 0 {
				sb.WriteByte(byte(0x00))
			}
			end = c - z
			v := s[start:end]
			logg.Tracef("explode", "page", i, "part start", start, "part end", end, "part str", v)
			v = s[start:end]
			sb.WriteString(v)
			start = end
		}
		lastPage = i
	}
	r := sb.String()
	r = strings.TrimRight(r, "\n")
	return r
}

//	if lastCursor <= capacity {
//		return pages, nil
//	}
//
//	var flatPages [][]uint32 
//
//	pages = append(pages, []uint32{})
//	for _, v := range bookmarks {
//		for _, vv := range v {
//			pages[0] = append(pages[0], vv)
//		}
//	}
//
//	var c uint32
//	var prev uint32
//	currentPage := 0 
//
//	for i, page := range pages {
//		var delta uint32
//		if i == 0 {
//			delta = nextSize
//		} else if i == len(pages) - 1 {
//			delta = prevSize
//		} else {
//			delta = nextSize + prevSize
//		}
//		remaining := capacity - delta
//		log.Printf("processing page %v", page)
//		lookAhead := page[1:]
//
//		for j, v := range lookAhead {
//			lc := v - c
//			log.Printf("currentpage j %v lc %v v %v remain %v", j, lc, v, remaining)
//			if lc > remaining {
//				c = prev
//				if lc - c > remaining {
//					return nil, fmt.Errorf("value at page %v idx %v alone exceeds capacity", i, v)
//				}
//				currentPage += 1
//				page = append(page, []uint32{})
//			}
//			page[currentPage] = append(page[currentPage], page[j])
//			prev = v
//		}
//	}
//	return page, nil
//}
