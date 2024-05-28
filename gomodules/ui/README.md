# UI package

this package takes inspiration from bubbletea and lipgloss to power some reusable components for your _non bubbletea_ cli application. This is useful when you have a cli application that you want to enhance but don't want to wrap all the logic in a bubbletea program. Also this package depends only on standards go package.

## Provided components
This are the components included in the package
### Confirm
```golang
confirm, err:= ui.NewConfirm("Please confirm", false).Run()
```
### Select
This uses a multiSelect component under the hood, already parameterized to handle a single selection.
```golang
options := []ui.SelectOption[int]{
  SelectOption[int]{Value: 1, Label: "option 1"},
  SelectOption[int]{Value: 2, Label: "option 2"},
  SelectOption[int]{Value: 3, Label: "option 3"},
}
selection, err := ui.NewSelect("Please select an option", options).
  MaxVisibleOptions(3).
  Run()

selection2, err := ui.NewSelectStrings("Please select an option from a list of strings", []{"option 1", "option 2", "option 3"}).
  Run()
```

#### Select Options
Here's a demo using all possible options
```golang
selection, err := ui.NewSelectStrings("Please select a fruit", []string{"apple", "cherry", "orange", "banana"}).
  SelectedIndex(2).
  MaxVisibleOptions(3).
  WithCleanup(true).
  Run()
```

### MultiSelect
This is a simple component to handle user selection it presents a menu to the user for selection it can handle multiple or single selection.
```golang
selection, err := ui.NewSelect("Choose an option", options).Run()

```
#### MultiSelect Options
Here's a demo using some possible options
```golang
selection, err := ui.NewSelect("Choose an option", options).
  SelectionMaxLen(2).              // default to 0 which means no limit
  SelectionMinLen(2).              // default to 1
  SelectionExactLen(2).            // this does exactly the same as the two lines above
  AllowEmptySelection().           // this does the same as SelectionMinLen(0)
  // following options are ignored in fallback mode
  SelectedIndexed(1,2).            // set a pre-selection of selected item
  MaxVisibleOptions(5).            // default value can be changed with ui.SetDefaultMaxVisibleOptions(5)
  WithCleanup(true)                // remove the selector from output when done
  Run()

```

### InputText / InputPassword
A component to gather text input with current edition binding attached. It supports validation and auto-completion.
```golang
input, err := ui.NewInputText(msg)
```

#### InputText Options
```golang
inputStr, err := ui.NewInputText(msg).
	Inline(). // prompt will be appended to msg end without new line
	SetPrompt("👉 "). // set a fancy prompt
	SetMaxLen(100). // it will block input over the given number of character
	SetMaxWidth(10). // set the size of the input field
	WithValidator(func (val string) bool {
		// your validation logic goes here
	}).
	WithCleanUp(). // the msg and user input will be removed from display
	WithCompletion(func(startOfWord string, fullWord string) ([]string, error){
		// Receives the start of the word to complete and the full word which can be different
		// if cursor is placed within a word. 
		// Returns the list of available completions.
	}).
	Run()

```

#### Limitations
- Input password may display the password when in fallback mode
- If terminal width can't be detected it can lead to weird behavior, same goes if the terminal is resized during edition. 
- It doesn't support input of carriage return, line feed or tab, they all will be replaced by single space
- Escapes sequences will be removed


## How to create your own components
If you come from bubbletea, you will be familiar with most of what to do here. If not I encourage you to look at existing components for better understanding. Don't hesitate to contact me for more in depth knowledge.

### The ComponentApi
your component should implement the ModelInterface  which means it also has to implement a ```GetComponentApi() *ComponentApi``` method where component api looks like:
```golang
type ComponentApi struct {
	Done    bool // if true will stop to wait for user input
	Cleanup bool // if true will remove component from ouput when done
	InputReader inputReader // select user input type you want to listen to
}
```
manipulating this api from your model will tell the system when it's time to stop and if it should cleanup or refresh the view when done.


### keybindings helpers
In order to ease the use of key bindings and automatically generate the help string you should defined keybindings in your model Init method
like this:
```golang
func (m *model[T]) Init() ui.Cmd {
	m.bindings = NewKeyBindings[*model[T]]()
  // (only the first key specified will appear in the helper message)
	m.bindings.AddBinding("up,k", ui.Msgs["up"], func(m *model[T]) ui.Cmd {
		if m.focusedIndex > 0 {
			m.focusedIndex--
		}
		return nil
	})
	m.bindings.AddBinding("down,j", ui.Msgs["down"], func(m *model[T]) ui.Cmd {
		if m.focusedIndex < len(m.options)-1 {
			m.focusedIndex++
		}
		return nil
	})
  // if description is "" then the key binding won't appear in the helper message
	m.bindings.AddBinding("ctrl+c", "", func(m *model[T]) ui.Cmd {
    return ui.CmdKill
	})
	return nil
}
```
Reading this you should have noticed the use of the ui.Msgs instead of raw strings this is intended for consistency and to allow localization
of components. To localize components it's up to you to detect language and replace strings in ui.Msgs to sweets your needs (at least for now this may change in the future).
So you are encouraged to use predefined messages in your components when there's already one available for your use case.

Defining your keybindings this way will allow you to easily handle keys in your Update method and to generate an help message in your Render method like this:
```golang

func (m *model[T]) Update(msg ui.Msg) (ui.Model, ui.Cmd) {
	cmd := m.bindings.Handle(m, msg)
	if m.done {
		return m, ui.CmdQuit
	}
	return m, cmd
}

func (m *model[T]) Render() string {
	theme := GetTheme() // use theme defined method for consistency more on theme later
	var sb strings.Builder
	sb.WriteString(theme.Title(m.title) + "\n")
  // ... rest of your UI here ...//
  // add a help string (showing keyboard bindings)
  sb.WriteString(m.bindings.GetDescription())
  sb.WriteString("\n")
	return sb.String()
}
```

### the FallBack method
In addition to Init, Update, and Render components in this package MUST define a FallBack method. 
The Fallback method will be used when NO_COLOR or ASSIST env variables are set or when ```ui.ToggleEnhenced(false)``` have been called

This should be a less appealing version of your component, with no formatting and simple readline for user input. When used the Update method
will not be used, so your are on your own on the fallback method to handle inputs, and call the Fallback method again if you need to retry

here's a basic example: 
```golang
func (m *model[T]) Fallback() (ui.Model, error) {
	var sb strings.Builder
	// reset selected
	sb.WriteString(fmt.Sprintf("%s\n", m.title))
	if m.errorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error: %s\n", m.errorMsg))
		m.errorMsg = ""
	}
  // some methods are provided to assist you like: Readline, ReadInt, ReadInts
	ints, err := ReadInts(sb.String())+
	if err != nil {
		m.errorMsg = Msgs["notANumber"]
		return m.Fallback() // call the method again as we don't have a valid input
	}
	return m, nil
}
```

## Limitations
This is not intended to be used in asynchronous environment. Reading user input should be a synchronous tasks, you can read an interesting article on that topic [here](https://dr-knz.net/bubbletea-control-inversion.html)
