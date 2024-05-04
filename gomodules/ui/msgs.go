/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package ui

var Msgs = map[string]string{
	"up":                             "Move up",
	"down":                           "Move down",
	"select":                         "Select",
	"confirm":                        "Confirm",
	"cancel":                         "Cancel",
	"userCanceled":                   "Canceled by user",
	"limitReached":                   "Selection limit reached",
	"selectAllToggle":                "Select/Deselect all",
	"errorPrefix":                    "Error: ",
	"fallbackMultiSelectPrompt":      "Select items by entering corresponding numbers: (eg: 1 2 3) ",
	"fallbackSelectPrompt":           "Select one item entering corresponding number: ",
	"fallbackConfirmError":           "Invalid input, please enter 'y' or 'n'",
	"fallbackConfirmPromptTrue":      "[Y/n]: ",
	"fallbackConfirmPromptFalse":     "[y/N]: ",
	"fallbackConfirmHelpPromptTrue":  "type y/n then enter (default yes): ",
	"fallbackConfirmHelpPromptFalse": "type y/n then enter (default no): ",
	"notANumber":                     "Selection must only contain numbers",
	/* Must be formatted with a number */
	"outBoundMin": "You must select at least %d options",
	/* Must be formatted with a number */
	"outBoundMax": "You can not select more than %d options",
	/* Must be formatted with a number */
	"outOfRange": "Selection out of range please select numbers between 1 and %d",
}
