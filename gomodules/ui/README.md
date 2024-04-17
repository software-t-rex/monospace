# UI package

this package uses bubbletea and lipgloss to power some reusable components for your _non bubbletea_ cli application. This is useful when you have a cli application that you want to enhance but don't want to wrap all the logic in a bubbletea program.

## Provided components
This are the components included in the package
### Confirm
```golang
confirm := ui.NewConfirm("Please confirm", false).Run()
```
### Select
This uses a multiSelect component under the hood, already parameterized to handle a single selection.
```golang
options := []ui.SelectOption[int]{
  SelectOption[int]{Value: 1, Label: "option 1"},
  SelectOption[int]{Value: 2, Label: "option 2"},
  SelectOption[int]{Value: 3, Label: "option 3"},
}
selection := ui.NewSelect("Please select an option", options).
  MaxVisibleOptions(3).
  Run()

selection2 := ui.NewSelectStrings("Please select an option from a list of strings", []{"option 1", "option 2", "option 3"}).
  Run()
```

#### Select Options
Here's a demo using all possible options
```golang
selector := ui.NewSelectStrings("Please select a fruit", []string{"apple", "cherry", "orange", "banana"}).
  SelectedIndex(2).
  MaxVisibleOptions(3).
  WithCleanup(true).
  Run()
```

### MultiSelect
This is a simple component to handle user selection it presents a menu to the user for selection it can handle multiple or single selection.
```golang
selection := ui.NewSelect("Choose an option", options).Run()

```
#### MultiSelect Options
Here's a demo using all possible options
```golang
confirm := ui.NewSelect("Choose an option", options).
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

## How to create your own components
components are basically bubbletea application, so if you are already familiar with bubbletea you should already know most of what to do.

### keybindings
In order to ease the use of key bindings and automatically generate the help string you should defined keybindings in your model Init method
like this:
```golang
func (m *model[T]) Init() tea.Cmd {
	m.bindings = NewKeyBindings[*model[T]]()
  // (only the first key specified will appear in the helper message)
	m.bindings.AddBinding("up,k", ui.Msgs["up"], func(m *model[T]) tea.Cmd {
		if m.focusedIndex > 0 {
			m.focusedIndex--
		}
		return nil
	})
	m.bindings.AddBinding("down,j", ui.Msgs["down"], func(m *model[T]) tea.Cmd {
		if m.focusedIndex < len(m.options)-1 {
			m.focusedIndex++
		}
		return nil
	})
  // if description is "" then the key binding won't appear in the helper message
	m.bindings.AddBinding("ctrl+c", "", func(m *model[T]) tea.Cmd {
    return tea.Quit
	})
	return nil
}
```
Reading this you should have noticed the use of the ui.Msgs instead of raw strings this is intended for consistency and to allow localization
of components. To localize components it's up to you to detect language and replace strings in ui.Msgs to sweets your needs (at least for now this may change in the future).
So you are encouraged to use predefined messages in your components when there's already one available for your use case.

Defining your keybindings this way will allow you to easily handle keys in your Update method and to generate an help message in your View method like this:
```golang

func (m *model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := m.bindings.Handle(m, msg)
	if m.done {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *model[T]) View() string {
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
In addition to Init, Update, and View components in this package MUST define a FallBack method. 
The Fallback method will be used when NO_COLOR or ASSIST env variables are set or when ```ui.ToggleEnhenced(false)``` have been called

This should be a less appealing version of your component, with no formatting and simple readline for user input. When used the Update method
will not be used, so your are on your own on the fallback method to handle inputs, and call the Fallback method again if you need to retry

here's a basic example: 
```golang
func (m *model[T]) Fallback() ui.TeaModelWithFallback {
	var sb strings.Builder
	// reset selected
	sb.WriteString(fmt.Sprintf("%s\n", m.title))
	if m.errorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error: %s\n", m.errorMsg))
		m.errorMsg = ""
	}
  // some methods are provided to assist you like: Readline, ReadInt, ReadInts
	ints, err := ReadInts(sb.String())
	if err != nil {
		m.errorMsg = Msgs["notANumber"]
		return m.Fallback() // call the method again as we don't have a valid input
	}
	return m
}
```

## Limitations
This is not intended to be used in asynchronous environment. Reading user input should be a synchronous tasks, you can read an interesting article on that topic [here](https://dr-knz.net/bubbletea-control-inversion.html)
