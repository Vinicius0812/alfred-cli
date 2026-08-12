package cli

import (
	"fmt"
	"io"
	"strings"
)

func (a *application) runMenu(global globalOptions, msg messages) int {
	printBrand(a.stdout, msg)
	for {
		fmt.Fprintln(a.stdout, msg.text("menu_validate"))
		fmt.Fprintln(a.stdout, msg.text("menu_doctor"))
		fmt.Fprintln(a.stdout, msg.text("menu_init"))
		fmt.Fprintln(a.stdout, msg.text("menu_workflows"))
		fmt.Fprintln(a.stdout, msg.text("menu_run"))
		fmt.Fprintln(a.stdout, msg.text("menu_version"))
		fmt.Fprintln(a.stdout, msg.text("menu_exit"))
		fmt.Fprint(a.stdout, msg.text("menu_prompt"))

		choice, err := a.reader.ReadString('\n')
		if err != nil && len(choice) == 0 {
			if err == io.EOF {
				return exitSuccess
			}
			fmt.Fprintf(a.stderr, "alfred: %v\n", err)
			return exitFailure
		}
		switch strings.TrimSpace(choice) {
		case "0":
			fmt.Fprintln(a.stdout, msg.text("menu_goodbye"))
			return exitSuccess
		case "1":
			a.runConfig([]string{"validate"}, global, msg)
		case "2":
			a.runDoctor(nil, global, msg)
		case "3":
			a.runInit(nil, global, msg)
		case "4":
			a.runWorkflow([]string{"list"}, global, msg)
		case "5":
			fmt.Fprint(a.stdout, msg.text("menu_workflow_prompt"))
			workflowName, readErr := a.reader.ReadString('\n')
			if readErr != nil && len(workflowName) == 0 {
				return exitFailure
			}
			if strings.TrimSpace(workflowName) == "" {
				fmt.Fprintln(a.stderr, msg.text("run_usage"))
				continue
			}
			a.runWorkflowExecution([]string{strings.TrimSpace(workflowName)}, global, msg)
		case "6":
			fmt.Fprintf(a.stdout, "alfred %s\n", Version)
		default:
			fmt.Fprintln(a.stderr, msg.text("menu_invalid"))
		}
	}
}
