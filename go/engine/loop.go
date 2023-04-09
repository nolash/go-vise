package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
)

func Loop(startSym string, en *Engine, ctx context.Context) error {
	err := en.Init(startSym, ctx)
	if err != nil {
		return fmt.Errorf("cannot init: %v\n", err)
	}

	b := bytes.NewBuffer(nil)
	en.WriteResult(b)
	fmt.Println(b.String())

	running := true
	for running {
		reader := bufio.NewReader(os.Stdin)
		in, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("cannot read input: %v\n", err)
		}
		in = strings.TrimSpace(in)
		running, err = en.Exec([]byte(in), ctx)
		if err != nil {
			return fmt.Errorf("unexpected termination: %v\n", err)
		}
		b := bytes.NewBuffer(nil)
		en.WriteResult(b)
		fmt.Println(b.String())
	}
	return nil
}
