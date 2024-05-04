package main

import (
	"fmt"
	"os"

	"github.com/software-t-rex/monospace/gomodules/ui"
)

type failure struct {
	msg           ui.MsgKey
	expectedBytes string
	expectedDesc  string
}

func main() {
	term := ui.GetTerminal()
	theme := ui.GetTheme()

	failed := []failure{}
	success := 0
	for seq, desc := range ui.SequenceKeysMap {
		if desc == "ctrl+c" {
			continue
		}
		for {
			fmt.Print("Please test the following key sequence: " + desc)
			keymsg, err := ui.ReadKeyPressEvent(term)
			if err != nil {
				fmt.Printf(theme.Error("Error: %s"), err)
				return
			} else if keymsg.(ui.MsgKey).Value == "ctrl+c" {
				fmt.Println("Exiting test")
				os.Exit(0)
			}
			if keymsg.(ui.MsgKey).Value != desc {
				fmt.Printf("%s\r\n", theme.FailureIndicator())
				fmt.Printf(theme.Error("Expected: '%#v': '%#v', got: '%#v'\r\n"), seq, desc, keymsg.(ui.MsgKey).Value)
				if ui.ConfirmInline("Do you want to try again?", false) {
					continue
				}
				failed = append(failed, failure{msg: keymsg.(ui.MsgKey), expectedBytes: seq, expectedDesc: desc})
			} else if string(keymsg.(ui.MsgKey).ByteSeq) != seq {
				success++
				fmt.Printf("%s (wanted %#v, got %#v)\r\n", theme.SuccessIndicator(), seq, string(keymsg.(ui.MsgKey).ByteSeq))
			} else {
				success++
				fmt.Printf("%s\r\n", theme.SuccessIndicator())
			}
			break
		}
	}
	// report results
	fmt.Printf("Tested %d key sequences, %d failed\r\n", len(ui.SequenceKeysMap), len(failed))
	if len(failed) > 0 {
		fmt.Println("Failed sequences:")
		for _, failure := range failed {
			key := failure.msg
			fmt.Printf("Expected: '%#v': '%#v', got: '%#v'\r\n", failure.expectedBytes, failure.expectedDesc, key.Value)
		}
	}

}
