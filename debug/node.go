package debug

import (
	"fmt"
)

type Node struct {
	Name string
	Description string
	conn []string
	menus []string
}

var (
	NodeIndex = make(map[string]*Node)
)

func (n *Node) haveConn(peer string) bool {
	var v string
	for _, v = range(n.conn) {
		if peer == v {
			return true
		}
	}
	return false
}

func (n *Node) Connect(sym string) bool {
	logg.Infof("node connet is now", "node", fmt.Sprintf("%p", n), "name", n.Name, "conn", n.conn)
	peer := GetNode(sym)
	if n.haveConn(peer.Name) {
		return false
	}
	n.conn = append(n.conn, peer.Name)
	logg.Infof("node connet is after", "node", fmt.Sprintf("%p", n), "name", n.Name, "conn", n.conn)
	return true
}
 
func (n *Node) String() string {
	s := n.Name
	if len(n.conn) > 0 {
		s += fmt.Sprintf(", conn: %s", n.conn)
	}
	if len(n.menus) > 0 {
		s += fmt.Sprintf(", menu: %s", n.menus)
	}
	return s
}

func (n *Node) Next() *Node {
	logg.Infof("node is now", "node", fmt.Sprintf("%p", n), "name", n.Name, "conn", n.conn)
	if len(n.conn) == 0 {
		return nil
	}
	r := GetNode(n.conn[0])
	n.conn = n.conn[1:]
	logg.Infof("node is after", "node", fmt.Sprintf("%p", n), "name", n.Name, "conn", n.conn)
	return r
}

func (n *Node) MenuNext() string {
	if len(n.menus) == 0 {
		return ""
	}
	r := n.menus[0]
	n.menus = n.menus[1:]
	return r
}

func GetNode(sym string) *Node {
	var node Node
	r, ok := NodeIndex[sym]
	if !ok {
		node.Name = sym
		r = &node
		NodeIndex[sym] = &node
	}
	return r
}
