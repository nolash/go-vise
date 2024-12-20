package debug

type Node struct {
	Name string
	Description string
	conn []string
}

var (
	NodeIndex = make(map[string]Node)
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
