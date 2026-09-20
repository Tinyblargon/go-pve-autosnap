package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"go-pve-autosnap/cmd/autosnap"
	"go-pve-autosnap/internal/cli"
	"go-pve-autosnap/internal/cli/cron"
	"go-pve-autosnap/internal/cli/flag/format"
	"go-pve-autosnap/internal/filter"
	"go-pve-autosnap/internal/pool"
	"go-pve-autosnap/internal/proxmox"
	"go-pve-autosnap/internal/tracker"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
)

var (
	snapCmd = &cobra.Command{
		Use:   "summary",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			writer, err := cli.GetFlagOutput(cmd)
			if err != nil {
				return err
			}
			defer writer.Close()

			var s cli.Shared
			if s, err = cli.Setup(); err != nil {
				return err
			}
			snap := proxmox.Snapshot{
				Description: s.Description,
				NamePrefix:  s.Label,
				TimeFormat:  s.TimeFormat}
			if s.Cron == nil {
				return execute(s.Ctx, s.Client, writer, s.Filter, s.Pool, s.Format, snap)
			}
			return cron.Execute(s.Ctx, *s.Cron, func() error {
				return execute(s.Ctx, s.Client, writer, s.Filter, s.Pool, s.Format, snap)
			})
		}}
)

func init() {
	cli.RootCmd.AddCommand(snapCmd)
	cli.SetFlagOutput(snapCmd)
}

type jsonOutput struct {
	Summary jsonSummary `json:"summary"`
	Errors  []error     `json:"errors,omitempty"`
}

type jsonSummary struct {
	GuestsLxc     uint `json:"lxc_guests"`
	GuestsQemu    uint `json:"qemu_guests"`
	SnapshotsLxc  uint `json:"lxc_snapshots"`
	SnapshotsQemu uint `json:"qemu_snapshots"`
}

func execute(ctx context.Context, c pve.ClientNew, writer io.Writer, filter *filter.Filter, pool pool.Pool, outputFormat format.Enum, snap proxmox.Snapshot) error {
	summary := new(tracker.GuestSummary{})
	err := autosnap.Execute(ctx, c, filter, pool, func(ctx context.Context, c pve.ClientNew, vmr *pve.VmRef) error {
		snapshots, err := snap.ListSnapshots(ctx, c.Snapshot, vmr)
		if err != nil {
			return err
		}
		summary.Add(vmr.GetVmType(), snapshots)
		return nil
	})
	if err != nil {
		return err
	}

	guestsLxc, guestsQemu, snapshotsLxc, snapshotsQemu := summary.Return()

	switch outputFormat {
	case format.Cli:
		var b strings.Builder
		b.WriteString("lxc guests:     ")
		b.WriteString(strconv.FormatUint(uint64(guestsLxc), 10))
		b.WriteString("\n" + "qemu guests:    ")
		b.WriteString(strconv.FormatUint(uint64(guestsQemu), 10))
		b.WriteString("\n" + "lxc snapshots:  ")
		b.WriteString(strconv.FormatUint(uint64(snapshotsLxc), 10))
		b.WriteString("\n" + "qemu snapshots: ")
		b.WriteString(strconv.FormatUint(uint64(snapshotsQemu), 10))
		b.WriteByte('\n')
		writer.Write([]byte(b.String()))
	case format.Json, format.JsonPretty:
		jsonData := jsonOutput{Summary: jsonSummary{
			GuestsLxc:     guestsLxc,
			GuestsQemu:    guestsQemu,
			SnapshotsLxc:  snapshotsLxc,
			SnapshotsQemu: snapshotsQemu,
		}}
		switch outputFormat {
		case format.Json:
			out, _ := json.Marshal(jsonData)
			fmt.Fprint(writer, string(out))
		case format.JsonPretty:
			out, _ := json.MarshalIndent(jsonData, "", cli.JsonIndent)
			fmt.Fprintln(writer, string(out))
		}
	}
	return nil
}
