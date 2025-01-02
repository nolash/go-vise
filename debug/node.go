package debug

type Node struct {
	Name string
	Description string
	conn []string
}

var (
	NodeIndex = make(map[string]Node)
	MenuIndex = make(map[string]int)
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

func (n *Node) Connect(peer Node) bool {
	var ok bool
	_, ok = NodeIndex[peer.Name]
	if !ok {
		NodeIndex[peer.Name] = peer
	}
	if n.haveConn(peer.Name) {
		return false
	}
	n.conn = append(n.conn, peer.Name)
	return true
}

func (n *Node) String() string {
	return n.Name
}

func (n *Node) Next() *Node {
	if len(n.conn) == 0 {
		return nil
	}
	r := NodeIndex[n.conn[0]]
	n.conn = n.conn[1:]
	return &r
}

func AddMenu(s string) int {
	var ok bool
	_, ok = MenuIndex[s]
	if !ok {
		MenuIndex[s] = 0
	}
	MenuIndex[s] += 1
	return MenuIndex[s]
}
