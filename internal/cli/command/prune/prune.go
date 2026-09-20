package prune

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
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
	keep     int
	dryRun   bool
	pruneCmd = &cobra.Command{
		Use:   "prune",
		Short: "Prunes snapshots of the specified guests",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
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
	cli.RootCmd.AddCommand(pruneCmd)
	cli.FlagKeep(pruneCmd, &keep, 0)
	cli.FlagDryRun(pruneCmd, &dryRun)
}

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
				jsonData.Prune = append(jsonData.Prune, *a)
				mutex.Unlock()
			}
		}
	} else {
		f = func(o *proxmox.Output) {}
	}

	err := autosnap.Execute(ctx, c, filter, pool, func(ctx context.Context, c pve.ClientNew, vmr *pve.VmRef) error {
		out, err := snap.EnsureAmount(ctx, c.Snapshot, vmr)
		if err != nil {
			return err
		}
		f(out)
		return nil
	})
	if dryRun {
		switch outputFormat {
		case format.Json, format.JsonPretty:
			if err != nil {
				jsonData.Errors = cli.SetJsonError(err)
			}
			switch outputFormat {
			case format.Json:
				out, _ := json.Marshal(jsonData)
				fmt.Fprint(writer, string(out))
			case format.JsonPretty:
				out, _ := json.Marshal(jsonData, jsontext.WithIndent(cli.JsonIndent))
				fmt.Fprintln(writer, string(out))
			}
		}
	}
	return err
}

const whitespace = "   "

type jsonOutput struct {
	Prune  []proxmox.Output `json:"prune"`
	Errors []error          `json:"errors,omitempty"`
}
