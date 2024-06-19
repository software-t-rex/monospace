/*
Copyright © 2023 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2023 Jonathan Gotti <jgotti@jgotti.org>
*/

package tasks

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/gomodules/ui"
	"github.com/software-t-rex/monospace/gomodules/utils"
)

// display a list of tasks defined in the pipeline and let the user select one
// if cancelOption is given then the user can cancel the selection
func TaskSingleSelector(p Pipeline, prompt string, cancelOption string) (string, error) {
	if len(p) == 0 {
		return "", ErrNoAvailableOption
	}
	options := utils.MapGetKeys(p)
	slices.Sort(options)
	escapable := false
	if cancelOption != "" {
		options = append(options, cancelOption)
		escapable = true
	}
	selected, errSelTask := ui.NewSelectStrings(prompt, options).
		MaxVisibleOptions(10).
		Escapable(escapable).
		WithCleanup(true).
		Run()
	// user escape is ok when escapable is true and user select cancelOption
	if escapable {
		if (errSelTask != nil && errors.Is(errSelTask, ui.ErrSelectEscaped)) || (errSelTask == nil && selected == cancelOption) {
			return "", ui.ErrSelectEscaped
		}
	}
	return selected, errSelTask
}

type DependencySelectorOptions struct {
	Config    *app.MonospaceConfig
	Exclude   []string
	Selected  []string
	Escapable bool
}

// display a list of tasks that are dependables in the given pipeline
// and let the user select one or more of them
// if Escapable option is set to true will return options.Selected if user escaped
func DependencySelector(prompt string, options DependencySelectorOptions) ([]string, error) {
	pipeline, errPipeline := GetStandardizedPipeline(options.Config, false)
	if errPipeline != nil {
		return []string{}, errPipeline
	}
	dependables := pipeline.GetDependableTasks(options.Exclude, options.Config)
	if len(dependables) == 0 {
		return []string{}, ErrNoAvailableOption
	}
	slices.Sort(dependables)
	Selector := ui.NewMultiSelectStrings(prompt, dependables).
		MaxVisibleOptions(10).
		AllowEmptySelection().
		Escapable(options.Escapable).
		WithCleanup(true)
	if len(options.Selected) > 0 {
		// get selected indexes
		selectedIndexes := make([]int, len(options.Selected))
		for i, s := range options.Selected {
			index := utils.SliceFindIndex(dependables, s)
			if index != -1 {
				selectedIndexes[i] = index
			}
		}
		Selector.SelectedIndexes(selectedIndexes...)
	}
	selection, errSel := Selector.Run()
	if errSel != nil && errors.Is(errSel, ui.ErrSelectEscaped) {
		return options.Selected, errSel
	}
	return selection, errSel
}

func cutString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func cycleOutputMode(task *app.MonospaceConfigTask, direction int) {
	outputmodes := []string{"", "grouped", "interleaved", "status-only", "errors-only", "none"}
	index := utils.SliceFindIndex(outputmodes, task.OutputMode)
	index += direction
	if index < 0 {
		index = len(outputmodes) - 1
	}
	if index >= len(outputmodes) {
		index = 0
	}
	task.OutputMode = outputmodes[index]
}

type TaskEditorSection int

const (
	// orders matters and must reflect the order in the render function
	TaskConfigSectionCmd TaskEditorSection = iota
	TaskConfigSectionDesc
	TaskConfigSectionDeps
	TaskConfigSectionPersist
	TaskConfigSectionOutputMode
	TaskConfigSectionSave
	TaskConfigSectionCancel
	// TaskConfigSectionEnv
)

var ErrEditorCanceled = errors.New("editor canceled")

type TaskEditorUIModel struct {
	task         Task
	originalTask Task
	bindings     *ui.KeyBindings[*TaskEditorUIModel]
	focusSection TaskEditorSection
	editing      bool
	uiAPI        *ui.ComponentApi
	config       *app.MonospaceConfig
	canceled     bool
}

func (m *TaskEditorUIModel) Init() ui.Cmd {
	m.bindings = ui.NewKeyBindings[*TaskEditorUIModel]()
	lastSection := TaskConfigSectionCancel
	m.bindings.
		AddBinding("enter", "", func(m *TaskEditorUIModel) ui.Cmd {
			switch m.focusSection {
			case TaskConfigSectionCmd, TaskConfigSectionDesc:
				if !m.editing {
					m.startLineEdition()
				}
			case TaskConfigSectionDeps:
				m.editing = true
			case TaskConfigSectionPersist:
				m.task.TaskDef.Persistent = !m.task.TaskDef.Persistent
			case TaskConfigSectionOutputMode:
				cycleOutputMode(&m.task.TaskDef, 1)
			case TaskConfigSectionSave:
				m.uiAPI.Done = true
			case TaskConfigSectionCancel:
				m.task = m.originalTask
				m.canceled = true
				m.uiAPI.Done = true
			}
			return nil
		}).
		AddBinding("up", "", func(m *TaskEditorUIModel) ui.Cmd {
			m.focusSection--
			if m.focusSection < TaskConfigSectionCmd {
				m.focusSection = lastSection
			}
			return nil
		}).
		AddBinding("down", "", func(m *TaskEditorUIModel) ui.Cmd {
			m.focusSection++
			if m.focusSection > lastSection {
				m.focusSection = TaskConfigSectionCmd
			}
			return nil
		}).
		AddBinding("left", "", func(m *TaskEditorUIModel) ui.Cmd {
			if m.focusSection == TaskConfigSectionOutputMode {
				cycleOutputMode(&m.task.TaskDef, -1)
			} else if m.focusSection == TaskConfigSectionPersist {
				m.task.TaskDef.Persistent = !m.task.TaskDef.Persistent
			}
			return nil
		}).
		AddBinding("right", "", func(m *TaskEditorUIModel) ui.Cmd {
			if m.focusSection == TaskConfigSectionOutputMode {
				cycleOutputMode(&m.task.TaskDef, 1)
			} else if m.focusSection == TaskConfigSectionPersist {
				m.task.TaskDef.Persistent = !m.task.TaskDef.Persistent
			}
			return nil
		}).
		AddBinding("ctrl+c", "", func(m *TaskEditorUIModel) ui.Cmd {
			m.uiAPI.Done = true
			return ui.CmdKill
		})
	return nil
}
func (m *TaskEditorUIModel) startLineEdition() {
	m.editing = true
	m.uiAPI.InputReader = ui.LineReader
}
func (m *TaskEditorUIModel) endLineEdition() ui.Msg {
	m.editing = false
	m.uiAPI.InputReader = ui.KeyReader
	return nil
}

func (m *TaskEditorUIModel) ReadlineConfig() ui.LineEditorOptions {
	var val string
	if m.editing {
		switch m.focusSection {
		case TaskConfigSectionCmd:
			val = strings.Join(m.task.TaskDef.Cmd, " ")
		case TaskConfigSectionDesc:
			val = m.task.TaskDef.Description
		}
	}
	return ui.LineEditorOptions{
		Value:     val,
		VisualPos: [2]int{0, 2},
	}
}

func (m *TaskEditorUIModel) ReadlineKeyHandler(key string) (ui.Msg, error) {
	switch key {
	case "esc":
		m.endLineEdition()
		return true, nil
	}
	return nil, nil
}

func (m *TaskEditorUIModel) Update(msg ui.Msg) ui.Cmd {
	if !m.editing {
		return m.bindings.Handle(m, msg)
	} else {
		switch msg := msg.(type) {
		case ui.MsgKey:
			key := msg.Value
			if key == "ctrl+c" {
				m.uiAPI.Done = true
				return ui.CmdKill
			}
		case ui.MsgLine:
			switch m.focusSection {
			case TaskConfigSectionCmd:
				m.task.TaskDef.Cmd = ParseArgs(msg.Value())
				m.endLineEdition()
			case TaskConfigSectionDesc:
				m.task.TaskDef.Description = msg.Value()
				m.endLineEdition()
			}
		}
	}
	return nil
}
func renderSection(theme *ui.Theme, title string, value string, isFocused bool, isEditing bool) string {
	if isFocused {
		title = theme.Bold(title)
		if isEditing {
			title = theme.Accentuated(title)
		}
	}
	return fmt.Sprintf("%s %s %s", theme.ConditionalFocusIndicator(isFocused), title, value)
}
func boldInFaint(s string) string {
	return fmt.Sprintf("\033[1m%s\033[22;2m", s)
}
func (m *TaskEditorUIModel) renderHelp(theme *ui.Theme) string {
	sb := strings.Builder{}
	keySep := theme.KeySeparator()
	keyBindSep := theme.KeyBindingSeparator()
	dfltMsg := fmt.Sprintf("%s to navigate %s %s to exit",
		boldInFaint(fmt.Sprintf("↑%s↓", keySep)),
		keyBindSep,
		boldInFaint("ctrl+c"),
	)

	switch m.focusSection {
	case TaskConfigSectionCmd:
		if m.editing {
			sb.WriteString(fmt.Sprintf("%s %s to save %s %s to cancel\n", theme.Accentuated(boldInFaint("Editing Cmd field:")), boldInFaint("↵"), keyBindSep, boldInFaint("esc")))
		} else {
			sb.WriteString(fmt.Sprintf("%s %s %s to edit\n", dfltMsg, keyBindSep, boldInFaint("↵")))
		}
	case TaskConfigSectionDesc:
		if m.editing {
			sb.WriteString(fmt.Sprintf("%s %s to save %s %s to cancel\n", theme.Accentuated(boldInFaint("Editing Description field:")), boldInFaint("↵"), keyBindSep, boldInFaint("esc")))
		} else {
			sb.WriteString(fmt.Sprintf("%s %s %s to edit\n", dfltMsg, keyBindSep, boldInFaint("↵")))
		}
	case TaskConfigSectionDeps:
		sb.WriteString(fmt.Sprintf("%s %s %s to edit\n", dfltMsg, keyBindSep, boldInFaint("↵")))
	case TaskConfigSectionPersist:
		sb.WriteString(fmt.Sprintf("%s %s %s to toggle\n", dfltMsg, keyBindSep, boldInFaint("↵")))
	case TaskConfigSectionOutputMode:
		sb.WriteString(fmt.Sprintf("%s %s %s to switch\n", dfltMsg, keyBindSep, boldInFaint(fmt.Sprintf("arrows%s↵", keySep))))
	case TaskConfigSectionSave:
		sb.WriteString(fmt.Sprintf("%s %s %s to exit\n", dfltMsg, keyBindSep, boldInFaint("↵")))
	default:
		sb.WriteString(fmt.Sprintf("%s\n", dfltMsg))
	}
	return ui.ApplyStyle(sb.String(), ui.Faint)
}
func (m *TaskEditorUIModel) Render() string {
	errorMsg := ""
	if m.focusSection == TaskConfigSectionDeps && m.editing {
		// this is a little bit hacky but the dependency selector will grab the rendering
		deps, errDeps := DependencySelector(fmt.Sprintf("Configuring %s\nSelect dependencies", m.task.Name.ConfigName), DependencySelectorOptions{
			Config:    m.config,
			Exclude:   []string{m.task.Name.String()},
			Selected:  m.task.TaskDef.DependsOn,
			Escapable: true,
		})
		if errDeps != nil && !errors.Is(errDeps, ui.ErrSelectEscaped) {
			errorMsg = errDeps.Error()
		}
		m.task.TaskDef.DependsOn = deps
		m.editing = false
	}
	theme := ui.GetTheme()
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(theme.Title("Configuring %s\n"), m.task.Name.ConfigName))
	sb.WriteString(renderSection(theme, "Cmd:", cutString(strings.Join(m.task.TaskDef.Cmd, " "), 40), m.focusSection == TaskConfigSectionCmd, m.editing))
	sb.WriteString("\n")
	sb.WriteString(renderSection(theme, "Description:", cutString(m.task.TaskDef.Description, 40), m.focusSection == TaskConfigSectionDesc, m.editing))
	sb.WriteString("\n")
	sb.WriteString(renderSection(theme, "Depends on:", "- "+strings.Join(m.task.TaskDef.DependsOn, "\n              - "), m.focusSection == TaskConfigSectionDeps, m.editing))
	sb.WriteString("\n")
	sb.WriteString(renderSection(theme, "Persistent:", fmt.Sprintf("%t", m.task.TaskDef.Persistent), m.focusSection == TaskConfigSectionPersist, m.editing))
	sb.WriteString("\n")
	sb.WriteString(renderSection(theme, "Output mode:", utils.If(m.task.TaskDef.OutputMode != "", m.task.TaskDef.OutputMode, fmt.Sprintf("default (%s)", m.config.PreferredOutputMode)), m.focusSection == TaskConfigSectionOutputMode, m.editing))
	sb.WriteString("\n")
	sb.WriteString(renderSection(theme, "Save", "", m.focusSection == TaskConfigSectionSave, m.editing))
	sb.WriteString("\n")
	sb.WriteString(renderSection(theme, "Cancel", "", m.focusSection == TaskConfigSectionCancel, m.editing))
	sb.WriteString("\n")
	sb.WriteString(m.renderHelp(theme))
	if errorMsg != "" {
		sb.WriteString(theme.Error(errorMsg))
		sb.WriteString("\n")
	}
	if m.editing {
		if m.focusSection == TaskConfigSectionCmd || m.focusSection == TaskConfigSectionDesc {
			sb.WriteString(theme.FocusItemIndicator())
		}
	}
	return sb.String()
}

func (m *TaskEditorUIModel) Fallback() (ui.Model, error) {
	fmt.Println("Fallback not implemented")
	return m, nil
}
func (m *TaskEditorUIModel) GetComponentApi() *ui.ComponentApi {
	return m.uiAPI
}

func (m *TaskEditorUIModel) Run() (Task, error) {
	model, err := ui.RunComponent(m)
	if err == nil && model.canceled {
		err = ErrEditorCanceled
	}
	return model.task, err
}

func NewTaskEditor(config *app.MonospaceConfig, task Task) *TaskEditorUIModel {
	return &TaskEditorUIModel{
		task: Task{
			Name: task.Name,
			TaskDef: app.MonospaceConfigTask{
				Description: task.TaskDef.Description,
				Cmd:         append([]string{}, task.TaskDef.Cmd...),
				DependsOn:   append([]string{}, task.TaskDef.DependsOn...),
				Persistent:  task.TaskDef.Persistent,
				OutputMode:  task.TaskDef.OutputMode,
			},
		},
		originalTask: task,
		focusSection: TaskConfigSectionCmd,
		config:       config,
		uiAPI:        &ui.ComponentApi{Cleanup: true},
	}
}
