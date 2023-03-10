package machine

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/client"
	"github.com/superfly/flyctl/flaps"
	"github.com/superfly/flyctl/internal/appconfig"
	"github.com/superfly/flyctl/internal/command"
	"github.com/superfly/flyctl/internal/flag"
	"github.com/superfly/flyctl/iostreams"
)

func newDestroy() *cobra.Command {
	const (
		short = "Destroy a Fly machine. This command requires a machine to be in a stopped state unless the force flag is used."
		long  = short + "\n"

		usage = "destroy <id>"
	)

	cmd := command.New(usage, short, long, runMachineDestroy,
		command.RequireSession,
		command.LoadAppNameIfPresent,
	)

	cmd.Aliases = []string{"remove", "rm"}

	flag.Add(
		cmd,
		flag.App(),
		flag.AppConfig(),
		flag.Bool{
			Name:        "force",
			Shorthand:   "f",
			Description: "force kill machine regardless of current state",
		},
	)

	cmd.Args = cobra.ExactArgs(1)

	return cmd
}

func runMachineDestroy(ctx context.Context) (err error) {
	var (
		out       = iostreams.FromContext(ctx).Out
		machineID = flag.FirstArg(ctx)
	)

	current, ctx, err := selectOneMachine(ctx, machineID)
	if err != nil {
		return err
	}
	flapsClient := flaps.FromContext(ctx)
	appName := appconfig.NameFromContext(ctx)

	// This is used for the deletion hook below.
	client := client.FromContext(ctx).API()
	app, err := client.GetAppCompact(ctx, appName)
	if err != nil {
		return fmt.Errorf("could not get app '%s': %w", appName, err)
	}

	if current.State != "stopped" {
		switch current.State {
		case "destroyed":
			return fmt.Errorf("machine %s has already been destroyed", current.ID)
		case "started":
			if !flag.GetBool(ctx, "force") {
				return fmt.Errorf("machine %s currently started, either stop first or use --force flag", current.ID)
			}
		default:
			return fmt.Errorf("machine %s is in a %s state and cannot be destroyed since it is not stopped, either stop first or use --force flag", current.ID, current.State)
		}
	}
	fmt.Fprintf(out, "machine %s was found and is currently in %s state, attempting to destroy...\n", current.ID, current.State)

	input := api.RemoveMachineInput{
		AppID: appName,
		ID:    current.ID,
		Kill:  flag.GetBool(ctx, "force"),
	}
	err = flapsClient.Destroy(ctx, input)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found") && appName != "":
			return fmt.Errorf("could not find machine %s in app %s to destroy", current.ID, appName)
		default:
			return fmt.Errorf("could not destroy machine %s: %w", current.ID, err)
		}
	}

	// Best effort post-deletion hook.
	runOnDeletionHook(ctx, app, current)

	fmt.Fprintf(out, "%s has been destroyed\n", current.ID)

	return
}
