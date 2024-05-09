package release

import (
	cmdCreate "github.com/jialequ/mplb/pkg/cmd/release/create"
	cmdDelete "github.com/jialequ/mplb/pkg/cmd/release/delete"
	cmdDeleteAsset "github.com/jialequ/mplb/pkg/cmd/release/delete-asset"
	cmdDownload "github.com/jialequ/mplb/pkg/cmd/release/download"
	cmdUpdate "github.com/jialequ/mplb/pkg/cmd/release/edit"
	cmdList "github.com/jialequ/mplb/pkg/cmd/release/list"
	cmdUpload "github.com/jialequ/mplb/pkg/cmd/release/upload"
	cmdView "github.com/jialequ/mplb/pkg/cmd/release/view"
	"github.com/jialequ/mplb/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdRelease(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release <command>",
		Short:   "Manage releases",
		GroupID: "core",
	}

	cmdutil.EnableRepoOverride(cmd, f)

	cmdutil.AddGroup(cmd, "General commands",
		cmdList.NewCmdList(f, nil),
		cmdCreate.NewCmdCreate(f, nil),
	)

	cmdutil.AddGroup(cmd, "Targeted commands",
		cmdView.NewCmdView(f, nil),
		cmdUpdate.NewCmdEdit(f, nil),
		cmdUpload.NewCmdUpload(f, nil),
		cmdDownload.NewCmdDownload(f, nil),
		cmdDelete.NewCmdDelete(f, nil),
		cmdDeleteAsset.NewCmdDeleteAsset(f, nil),
	)

	return cmd
}
