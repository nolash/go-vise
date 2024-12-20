package debug

import (
	"git.defalsify.org/vise.git/vm"
	"git.defalsify.org/vise.git/logging"
)

var (
	logg = logging.NewVanilla().WithDomain("debug")
)

type NodeParseHandler struct {
	*vm.ParseHandler
	node *Node
	parentMoveFunc func(string) error
	parentInCmpFunc func(string, string) error
}

func NewNodeParseHandler(node *Node) *NodeParseHandler {
	np := &NodeParseHandler{
		ParseHandler: vm.NewParseHandler().WithDefaultHandlers(),
	}
	np.node = node
	np.node.Name = node.Name
	np.parentMoveFunc = np.ParseHandler.Move
	np.parentInCmpFunc = np.ParseHandler.InCmp
	np.Move = np.move
	np.InCmp = np.incmp
	return np
}

func (np *NodeParseHandler) move(sym string) error {
	var node Node

	if (sym == "<" || sym == ">" || sym == "^" || sym == "_") {
		logg.Debugf("skip lateral move")
		return np.parentMoveFunc(sym)
	}

	node.Name = sym
	np.node.Connect(node)
	logg.Infof("connect MOVE", "src", np.node.Name, "dst", node.Name)
	return np.parentMoveFunc(sym)
}

func (np *NodeParseHandler) incmp(sym string, sel string) error {
	var node Node

	if (sym == "<" || sym == ">" || sym == "^" || sym == "_") {
		logg.Debugf("skip relative move")
		return np.parentMoveFunc(sym)
	}


	node.Name = sym
	np.node.Connect(node)
	logg.Debugf("connect INCMP", "src", np.node.Name, "dst", node.Name)
	return np.parentInCmpFunc(sym, sel)
}
