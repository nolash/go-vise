package debug

import (
	"fmt"

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
	parentCatchFunc func(string, uint32, bool) error
	parentMOutFunc func(string, string) error
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
	//np.parentMOutFunc = np.ParseHandler.MOut
	np.Move = np.move
	np.InCmp = np.incmp
	np.Catch = np.catch
	//np.MOut = np.mout
	return np
}

func (np *NodeParseHandler) mout(name string, sel string) error {
	np.node.menus = append(np.node.menus, name)
	logg.Infof("add MOUT", "src", np.node.Name, "name", name, "node", fmt.Sprintf("%p", np.node))
	return np.parentMOutFunc(name, sel)
}

func (np *NodeParseHandler) move(sym string) error {
	if (sym == "<" || sym == ">" || sym == "^" || sym == "_" || sym == ".") {
		logg.Debugf("skip lateral move")
		return np.parentMoveFunc(sym)
	}

	np.node.Connect(sym)
	logg.Infof("connect MOVE", "src", np.node.Name, "dst", sym)
	return np.parentMoveFunc(sym)
}

func (np *NodeParseHandler) incmp(sym string, sel string) error {
	if (sym == "<" || sym == ">" || sym == "^" || sym == "_" || sym == ".") {
		logg.Debugf("skip relative move")
		return np.parentInCmpFunc(sym, sel)
	}


	np.node.Connect(sym)
	logg.Debugf("connect INCMP", "src", np.node.Name, "dst", sym)
	return np.parentInCmpFunc(sym, sel)
}

func (np *NodeParseHandler) catch(sym string, flag uint32, inv bool) error {
	if (sym == "<" || sym == ">" || sym == "^" || sym == "_" || sym == ".") {
		logg.Debugf("skip relative move")
		return np.parentMoveFunc(sym)
	}

	np.node.Connect(sym)
	logg.Debugf("connect CATCH", "src", np.node.Name, "dst", sym)
	return np.parentCatchFunc(sym, flag, inv)
}
