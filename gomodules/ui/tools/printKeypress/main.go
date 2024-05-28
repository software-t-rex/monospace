package main

import (
	"errors"
	"fmt"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func main() {
	term := ui.GetTerminal()
	theme := ui.GetTheme()

	for {
		key, err := ui.ReadKeyPressEvent(term)
		if err != nil && !errors.Is(err, ui.ErrUnknownKey) {
			fmt.Printf("Error: %s\r\n", err)
			break
		}
		switch key := key.(type) {
		case ui.MsgKey:
			if !key.Unknown {
				fmt.Printf("Key: '%v', isEscapeSeq: %t, seq:%#v \r\n", key.Value, key.IsSeq, string(key.ByteSeq))
			} else {
				fmt.Printf(theme.Error("Key: '%v', isEscapeSeq: %t, seq:%#v \r\n"), key.Value, key.IsSeq, string(key.ByteSeq))
				fmt.Printf(theme.Info("To support this sequence, add it to ui/pkg/sequencesKeys/sequenceKeys.go:")+"\r\n\t%#v: \"KEYDESCRIPTION\", \r\n", string(key.ByteSeq))
			}
			if key.Value == "ctrl+c" {
				return
			}
		}
	}
}
