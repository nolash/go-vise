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
	parentMOutFunc func(string, string) error
	parentMoveFunc func(string) error
	parentInCmpFunc func(string, string) error
	parentCatchFunc func(string, uint32, bool) error
}

func NewNodeParseHandler(node *Node) *NodeParseHandler {
	np := &NodeParseHandler{
		ParseHandler: vm.NewParseHandler().WithDefaultHandlers(),
	}
	np.node = node
	np.node.Name = node.Name
	np.parentMoveFunc = np.ParseHandler.Move
	np.parentInCmpFunc = np.ParseHandler.InCmp
	np.parentCatchFunc = np.ParseHandler.Catch
	np.parentMOutFunc = np.ParseHandler.MOut
	np.Move = np.move
	np.InCmp = np.incmp
	np.Catch = np.catch
	np.MOut = np.mout
	return np
}

func (np *NodeParseHandler) mout(sym string, sel string) error {
	c := AddMenu(sym)
	logg.Infof("add MOUT", "sym", sym, "visited", c)
	return np.parentMOutFunc(sym, sel)
}

func (np *NodeParseHandler) move(sym string) error {
	var node Node

	if (sym == "<" || sym == ">" || sym == "^" || sym == "_" || sym == ".") {
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

	if (sym == "<" || sym == ">" || sym == "^" || sym == "_" || sym == ".") {
		logg.Debugf("skip relative move")
		return np.parentInCmpFunc(sym, sel)
	}


	node.Name = sym
	np.node.Connect(node)
	logg.Debugf("connect INCMP", "src", np.node.Name, "dst", node.Name)
	return np.parentInCmpFunc(sym, sel)
}

func (np *NodeParseHandler) catch(sym string, flag uint32, inv bool) error {
	var node Node

	if (sym == "<" || sym == ">" || sym == "^" || sym == "_" || sym == ".") {
		logg.Debugf("skip relative move")
		return np.parentMoveFunc(sym)
	}

	node.Name = sym
	np.node.Connect(node)
	logg.Debugf("connect CATCH", "src", np.node.Name, "dst", node.Name)
	return np.parentCatchFunc(sym, flag, inv)
}
