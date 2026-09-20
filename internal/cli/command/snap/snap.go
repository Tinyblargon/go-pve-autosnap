package snap

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"go-pve-autosnap/cmd/autosnap"
	"go-pve-autosnap/internal/cli"
	"go-pve-autosnap/internal/cli/cron"
	"go-pve-autosnap/internal/cli/flag/format"
	"go-pve-autosnap/internal/filter"
	"go-pve-autosnap/internal/pool"
	"go-pve-autosnap/internal/proxmox"
	"io"
	"os"
	"strings"
	"sync"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var (
	keep    int
	state   bool
	dryRun  bool
	snapCmd = &cobra.Command{
		Use:   "snap",
		Short: "Creates a new snapshot of the specified guests",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if keep == 0 {
				return errors.New("--keep may not be 0")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := cli.Setup()
			if err != nil {
				return err
			}
			snap := proxmox.Snapshot{
				AmountToKeep: keep,
				Description:  s.Description,
				DryRun:       dryRun,
				NamePrefix:   s.Label,
				TimeFormat:   s.TimeFormat}
			if s.Cron == nil {
				return execute(s.Ctx, s.Client, os.Stdout, s.Filter, s.Pool, s.Format, snap)
			}
			if dryRun {
				if err = cron.Validate(*s.Cron); err != nil {
					return err
				}
				return execute(s.Ctx, s.Client, os.Stdout, s.Filter, s.Pool, s.Format, snap)
			}
			return cron.Execute(s.Ctx, *s.Cron, func() error {
				return execute(s.Ctx, s.Client, os.Stdout, s.Filter, s.Pool, s.Format, snap)
			})
		}}
)

func init() {
	cli.RootCmd.AddCommand(snapCmd)
	cli.FlagDryRun(snapCmd, &dryRun)
	cli.FlagKeep(snapCmd, &keep, -1)
	snapCmd.Flags().BoolVarP(&state, cli.FlagState, "s", false, "whether memory state should be included in the snapshot")
}

const whitespace = "   "

func execute(ctx context.Context, c pve.ClientNew, writer io.Writer, filter *filter.Filter, pool pool.Pool, outputFormat format.Enum, snap proxmox.Snapshot) error {

	var f func(*proxmox.Output)
	var jsonData jsonOutput
	if dryRun {
		switch outputFormat {
		case format.Cli:
			f = func(a *proxmox.Output) {
				var b strings.Builder
				b.WriteString("guest: ")
				b.WriteString(a.GuestType.String())
				b.WriteByte('/')
				b.WriteString(a.GuestID.String())
				b.WriteString("\n" + whitespace + "create snapshot: ")
				b.WriteString(a.Create.String())
				b.WriteByte('\n')
				if len(a.Remove) > 0 {
					b.WriteString(whitespace + "prune snapshots:\n")
					for i := range a.Remove {
						b.WriteString(whitespace + whitespace)
						b.WriteString(a.Remove[i].String())
						b.WriteByte('\n')
					}
				}
				writer.Write([]byte(b.String()))
			}
		case format.Json, format.JsonPretty:
			var mutex sync.Mutex
			f = func(a *proxmox.Output) {
				mutex.Lock()
				jsonData.Snap = append(jsonData.Snap, *a)
				mutex.Unlock()
			}
		}
	} else {
		f = func(o *proxmox.Output) {}
	}

	err := autosnap.Execute(ctx, c, filter, pool, func(ctx context.Context, c pve.ClientNew, vmr *pve.VmRef) error {
		out, err := snap.AddAndRemove(ctx, c, vmr)
		if err != nil {
			return err
		}
		f(out)
		return nil
	})
	if err != nil {
		return err
	}
	if dryRun {
		switch outputFormat {
		case format.Json:
			out, _ := json.Marshal(jsonData)
			fmt.Fprint(writer, string(out))
		case format.JsonPretty:
			out, _ := json.Marshal(jsonData, jsontext.WithIndent(cli.JsonIndent))
			fmt.Fprintln(writer, string(out))
		}
	}
	return nil
}

type jsonOutput struct {
	Snap   []proxmox.Output `json:"snap"`
	Errors []error          `json:"errors,omitempty"`
}
