package scale

import (
	"context"
	"fmt"

	"github.com/superfly/flyctl/client"
	"github.com/superfly/flyctl/internal/appconfig"
	"github.com/superfly/flyctl/internal/command"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	const (
		short = "Scale app resources"
		long  = `Scale application resources`
	)
	cmd := command.New("scale", short, long, nil)
	cmd.AddCommand(
		newScaleVm(),
		newScaleMemory(),
		newScaleShow(),
		newScaleCount(),
	)
	return cmd
}

func failOnMachinesApp(ctx context.Context) (context.Context, error) {
	apiClient := client.FromContext(ctx).API()
	appName := appconfig.NameFromContext(ctx)

	app, err := apiClient.GetAppBasic(ctx, appName)
	if err != nil {
		return nil, err
	} else if app.PlatformVersion == appconfig.MachinesPlatform {
		return nil, fmt.Errorf("This command doesn't support V2 apps yet, use `fly machines update` and `fly machines clone` instead")
	}

	return ctx, nil
}
