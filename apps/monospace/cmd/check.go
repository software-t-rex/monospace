/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/software-t-rex/go-jobExecutor/v2"
	"github.com/software-t-rex/monospace/utils"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckConfigFound(true)
		// monoRoot := utils.MonospaceGetRoot()
		// config := utils.CheckErrOrReturn(app.ConfigGet())
		utils.CheckErr(utils.MonospaceChdir())
		outputMode := ValidateFlagOutputMode(cmd)
		filteredProjects := GetFilterProjects(cmd, true)
		monoOrigin := utils.CheckErrOrReturn(utils.GitGetOrigin("./"))
		checkRepo := func(project utils.Project) func() (string, error) {
			return func() (string, error) {
				// check the directory exists
				if isdir, _ := utils.IsDir(project.Name); !isdir {
					advice := ""
					err := fmt.Errorf("directory %s does not exist", project.Name)
					if project.Kind == utils.Local {
						advice = "should init a local repository"
					} else {
						advice = fmt.Sprintf("Should clone %s to %s", project.RepoUrl, project.Name)
					}
					return "", fmt.Errorf("%w %s", err, advice)
				}
				// if this is an external project check it point to the correct repo
				origin, err := utils.GitGetOrigin(project.Name)
				if err != nil {
					err = fmt.Errorf("error checking %s origin: %w", project.Name, err)
					return "", err
				} else if project.RepoUrl != origin {
					err := fmt.Errorf("%s mismatch origin for %s!=%s: should check manually", project.Name, origin, project.RepoUrl)
					return "", err
				}
				return "ok", nil
			}
		}
		e := utils.NewTaskExecutor(outputMode)
		e.AddJob(jobExecutor.NamedJob{Name: "monospace", Job: checkRepo(utils.Project{Name: "./", Kind: utils.Internal, RepoUrl: monoOrigin})})
		for _, project := range filteredProjects {
			e.AddJob(jobExecutor.NamedJob{Name: project.Name, Job: checkRepo(project)})
		}
		e.Execute()
	},
}

func init() {
	repoCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
