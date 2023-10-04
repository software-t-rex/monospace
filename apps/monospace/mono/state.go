package mono

import (
	"os"
	"path/filepath"

	"github.com/software-t-rex/monospace/app"
	"github.com/software-t-rex/monospace/git"
	"github.com/software-t-rex/monospace/utils"
	"gopkg.in/yaml.v3"
)

type MonospaceState struct {
	Project  string `yaml:"project"`
	Revision string `yaml:"rev"`
}

type MonospaceStateList struct {
	States map[string][]MonospaceState `yaml:"states"`
}

const stateFile = ".pinnedStates.yml"

var cachedStates MonospaceStateList

func StateSave(states MonospaceStateList) error {
	raw, err := yaml.Marshal(states)
	config := utils.CheckErrOrReturn(app.ConfigGet())
	filePath := filepath.Join(config.GetDir(), stateFile)
	if err != nil {
		return err
	}
	raw = append([]byte("# You should not edit this file manually. It can corrupt your pinned states.\n"), raw...)
	return os.WriteFile(filePath, raw, 0640)
}

func StateLoadNoCache() (MonospaceStateList, error) {
	states := MonospaceStateList{}
	config := utils.CheckErrOrReturn(app.ConfigGet())
	filePath := filepath.Join(config.GetDir(), stateFile)
	if !utils.FileExistsNoErr(filePath) {
		return states, nil
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return states, err
	}
	err = yaml.Unmarshal(raw, &states)
	return states, err
}

func StateLoad() (MonospaceStateList, error) {
	if cachedStates.States == nil {
		var err error
		cachedStates, err = StateLoadNoCache()
		if err != nil {
			return cachedStates, err
		}
	}
	return cachedStates, nil
}

func StateList() []string {
	states, err := StateLoad()
	if err != nil {
		utils.Exit(err.Error())
	}
	names := []string{}
	for name := range states.States {
		names = append(names, name)
	}
	return names
}
func (s *MonospaceStateList) Save() error {
	return StateSave(*s)
}
func (s *MonospaceStateList) Len() (res int) {
	if s.States == nil {
		return 0
	}
	return len(s.States)
}
func (s *MonospaceStateList) Add(name string) {
	// first check there is no state with the same name
	if _, exists := s.States[name]; exists {
		utils.Exit("state " + name + " already exists.")
	}
	// init the state map if needed
	if s.States == nil {
		s.States = make(map[string][]MonospaceState)
	}
	// add monospace root state
	revRoot, errRoot := git.GetRevision(SpaceGetRoot())
	if errRoot != nil {
		utils.Exit(errRoot.Error())
	}
	s.States[name] = []MonospaceState{{Project: "root", Revision: revRoot}}
	// for each projects in the monospace, get the current revision
	projects := ProjectsGetAll()
	for _, p := range projects {
		if p.Kind != External {
			continue
		}
		rev, err := git.GetRevision(p.Path())
		if err != nil {
			utils.Exit(err.Error())
		}
		s.States[name] = append(s.States[name], MonospaceState{Project: p.Name, Revision: rev})
	}
}

func (s *MonospaceStateList) Remove(name string) {
	if _, exists := s.States[name]; !exists {
		utils.Exit("state " + name + " doesn't exists.")
	}
	delete(s.States, name)
}

func (s *MonospaceStateList) Restore(name string) {
	if _, exists := s.States[name]; !exists {
		utils.Exit("state " + name + " doesn't exists.")
	}
	uncleanProjects := []string{}
	unknownProjects := []string{}
	notGitProjects := []string{}
	// first check that all projects are in a clean state
	if !git.IsClean(SpaceGetRoot(), "") {
		uncleanProjects = append(uncleanProjects, "root")
	}
	for _, state := range s.States[name][1:] {
		if !git.IsClean(ProjectGetPath(state.Project), "") {
			uncleanProjects = append(uncleanProjects, state.Project)
		}
	}
	// then check all projects in states are part of the monospace and are git projects
	for _, state := range s.States[name][1:] {
		if !ProjectExists(state.Project) {
			unknownProjects = append(unknownProjects, state.Project)
		} else if !utils.FileExistsNoErr(filepath.Join(ProjectGetPath(state.Project), ".git")) {
			notGitProjects = append(notGitProjects, state.Project)
		}
	}
	// if there are unclean projects present a list of unclean projects and ask for confirmation
	if len(uncleanProjects) > 0 {
		for _, p := range uncleanProjects {
			utils.PrintWarning(p + " is not in a clean state")
		}
		if !utils.Confirm("Some projects are not in a clean state. Are you sure you want to continue ?", false) {
			utils.Exit("Aborted")
		}
	}
	// if there are unknown projects display the list of unknwon projects ans ask if user want to continue
	if len(unknownProjects) > 0 {
		for _, p := range unknownProjects {
			utils.PrintWarning(p + " is not part of the monospace")
		}
		if !utils.Confirm("Some projects in pinned state are not part of the monospace and won't be restored. Are you sure you want to continue ?", false) {
			utils.Exit("Aborted")
		}
	}
	// if there are not git projects display the list of not git projects ans ask if user want to continue
	if len(notGitProjects) > 0 {
		for _, p := range notGitProjects {
			utils.PrintWarning(p + " is not a git project")
		}
		if !utils.Confirm("Some projects in pinned state are not git projects and won't be restored. Are you sure you want to continue ?", false) {
			utils.Exit("Aborted")
		}
	}
	// now for all projects in states that are not in uncleanProjects, unknownProjects or notGitProjects restore them to pinned state
	utils.CheckErr(git.CheckoutRev(SpaceGetRoot(), s.States[name][0].Revision))
	for _, state := range s.States[name][1:] {
		if !utils.SliceContains(uncleanProjects, state.Project) && !utils.SliceContains(unknownProjects, state.Project) && !utils.SliceContains(notGitProjects, state.Project) {
			err := git.CheckoutRev(ProjectGetPath(state.Project), state.Revision)
			if err != nil {
				utils.PrintWarning(err.Error())
				utils.PrintWarning("Failed to restore " + state.Project + " to revision " + state.Revision)
			}
		}
	}
}
