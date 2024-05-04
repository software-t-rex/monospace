package main

import (
	"fmt"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

func main() {
	term := ui.GetTerminal()
	theme := ui.GetTheme()

	for {
		key, err := ui.ReadKeyPressEvent(term)
		if err != nil {
			fmt.Printf("Error: %s\r\n", err)
			break
		}
		switch key := key.(type) {
		case ui.MsgKey:
			if !key.Unknown {
				fmt.Printf("Key: '%+v', isEscapeSeq: %t, seq:%#v \r\n", key.Value, key.IsSeq, string(key.ByteSeq))
			} else {
				fmt.Printf(theme.Error("Key: '%#v', isEscapeSeq: %t, seq:%#v \r\n"), key.Value, key.IsSeq, string(key.ByteSeq))
				fmt.Printf("To add support add it to the list of known keys in ui/keys.go:\t\"%#v\": \"KEYDESCRIPTION\", \r\n", string(key.ByteSeq))
			}
			if key.Value == "ctrl+c" {
				return
			}
		}
	}
}
