package cli

import (
	"context"
	"crypto/tls"
	"errors"
	"go-pve-autosnap/internal/cli/cron"
	"go-pve-autosnap/internal/cli/flag/datetime"
	"go-pve-autosnap/internal/cli/flag/format"
	"go-pve-autosnap/internal/filter"
	"go-pve-autosnap/internal/pool"
	"go-pve-autosnap/internal/proxmox"
	"os"
	"time"

	pve "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "go-pve-autosnap",
		Short: "Application to make snapshots of guest systems inside of Proxmox",
	}
	outputFormat format.Enum
)

const (
	JsonIndent = "    "
)

const (
	flagDryRun          = "dry-run"
	flagKeep            = "keep"
	FlagState           = "state"
	flagDescription     = "description"
	flagFilter          = "filter"
	flagFormat          = "format"
	flagInsecure        = "insecure"
	flagLabel           = "label"
	flagMaxParallel     = "max-parallel"
	flagMaxParallelNode = "max-parallel-node"
	FlagOutput          = "output"
	flagPassword        = "password"
	flagPasswordFile    = "password-file"
	flagTimeout         = "timeout"
	flagToken           = "token"
	flagTokenFile       = "token-file"
	flagURL             = "url"
	flagUsername        = "username"
	flagUsernameFile    = "username-file"
)

func init() {
	RootCmd.PersistentFlags().StringP(flagURL, "u", "", "PVE API url")
	if err := RootCmd.MarkPersistentFlagRequired(flagURL); err != nil {
		panic(err)
	}
	RootCmd.PersistentFlags().BoolP(flagInsecure, "i", false, "ignore HTTPS host verification")

	RootCmd.PersistentFlags().StringP(flagToken, "t", "", "api token")
	RootCmd.PersistentFlags().StringP(flagTokenFile, "T", "", "file containing the api token")
	RootCmd.MarkFlagFilename(flagTokenFile)
	RootCmd.MarkFlagsMutuallyExclusive(flagToken, flagTokenFile)
	RootCmd.PersistentFlags().StringP(flagUsername, "n", "", "username")
	RootCmd.PersistentFlags().StringP(flagUsernameFile, "N", "", "file containing the username")
	RootCmd.MarkFlagFilename(flagUsernameFile)
	RootCmd.MarkFlagsMutuallyExclusive(flagUsername, flagUsernameFile, flagToken, flagTokenFile)
	RootCmd.PersistentFlags().StringP(flagPassword, "p", "", "password")
	RootCmd.PersistentFlags().StringP(flagPasswordFile, "P", "", "file containing the password")
	RootCmd.MarkFlagFilename(flagPasswordFile)
	RootCmd.MarkFlagsMutuallyExclusive(flagPassword, flagPasswordFile, flagToken, flagTokenFile)
	RootCmd.MarkFlagsOneRequired(flagUsername, flagUsernameFile, flagPassword, flagPasswordFile, flagToken, flagTokenFile)

	RootCmd.PersistentFlags().StringP(flagFilter, "f", "@all:", "filter expression to selects guests to process")
	RootCmd.PersistentFlags().StringP(flagDescription, "d", "go-pve-autosnap", "description of the snapshot, used for filtering snapshots created by this application")
	RootCmd.PersistentFlags().Uint(flagTimeout, 5, "timeout in seconds for the connection to the PVE server")
	RootCmd.PersistentFlags().StringP(cron.Flag, "c", "", "when set the application wil run in daemon mode and execute the command according to the schedule in cron format")
	RootCmd.PersistentFlags().Uint(flagMaxParallel, 1, "Maximum number of guests to process in parallel.")
	RootCmd.PersistentFlags().Uint(flagMaxParallelNode, 0, "Maximum number of guests to process in parallel per node")
	RootCmd.PersistentFlags().StringP(flagLabel, "l", "auto-", "prefix for the snapshot name, the time will be appended to it")
	RootCmd.PersistentFlags().String(datetime.Flag, "YYYYMMDDhhmmss", "datetime format for the snapshot name")
	RootCmd.PersistentFlags().Var(&outputFormat, flagFormat, `output format, allowed: "`+format.CliString+`", "`+format.JsonString+`", "`+format.JsonPrettyString+`"`)

	RootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if GetFlagDescription() == "" {
			return errors.New("--" + flagDescription + " may not be empty")
		}
		if GetFlagLabel() == "" {
			return errors.New("--" + flagLabel + " may not be empty")
		}
		if GetFlagMaxParallel() == 0 {
			return errors.New("--" + flagMaxParallel + " may not be 0")
		}
		if v := GetFlagMaxParallelNode(); v != nil && *v == 0 {
			return errors.New("--" + flagMaxParallelNode + " may not be 0")
		}

		return nil
	}
}

func Execute(version string) (err error) {
	RootCmd.Version = version
	RootCmd.SetVersionTemplate("{{.Version}}\n")
	return RootCmd.Execute()
}

func NewFilterAndClient(ctx context.Context) (filterSteps *filter.Filter, client pve.ClientNew, err error) {
	filterSteps, err = newFilter()
	if err != nil {
		return
	}
	client, err = newClient(ctx)
	return
}

func newFilter() (*filter.Filter, error) {
	rawFilter, _ := RootCmd.Flags().GetString(flagFilter)
	filters := filter.New()
	filters.Register(filter.ConstructorAll())
	filters.Register(filter.ConstructorID())
	filters.Register(filter.ConstructorName())
	filters.Register(filter.ConstructorNode())
	filters.Register(filter.ConstructorTag())
	return filters.Parse(rawFilter)
}

func newClient(ctx context.Context) (pve.ClientNew, error) {
	insecure, _ := RootCmd.Flags().GetBool(flagInsecure)
	timeout, _ := RootCmd.Flags().GetUint(flagTimeout)
	url, _ := RootCmd.Flags().GetString(flagURL)

	var username, password string
	var tokenAuth bool
	tokenRaw, _ := RootCmd.Flags().GetString(flagToken)
	tokenFile, _ := RootCmd.Flags().GetString(flagTokenFile)
	if tokenRaw != "" || tokenFile != "" {
		tokenAuth = true
		if tokenFile != "" {
			v, err := os.ReadFile(tokenFile)
			if err != nil {
				return pve.ClientNew{}, err
			}
			tokenRaw = string(v)
		}
	} else {
		username, _ = RootCmd.Flags().GetString(flagUsername)
		if file, _ := RootCmd.Flags().GetString(flagUsernameFile); file != "" {
			v, err := os.ReadFile(file)
			if err != nil {
				return pve.ClientNew{}, err
			}
			username = string(v)
		}
		password, _ = RootCmd.Flags().GetString(flagPassword)
		if file, _ := RootCmd.Flags().GetString(flagPasswordFile); file != "" {
			v, err := os.ReadFile(file)
			if err != nil {
				return pve.ClientNew{}, err
			}
			password = string(v)
		}
	}

	var tlsConf *tls.Config
	if insecure {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}
	c, err := pve.NewClient(url, nil, "", tlsConf, "", int(timeout), false)
	if err != nil {
		return pve.ClientNew{}, err
	}
	tmpCTX, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if tokenAuth {
		var token pve.ApiToken
		if err = token.Parse(tokenRaw); err != nil {
			return pve.ClientNew{}, err
		}
		c.SetAPIToken(token)
		// As test, get the version of the server
		_, err = c.GetVersion(tmpCTX)
	} else {
		err = c.Login(tmpCTX, username, password, "")
	}
	if err != nil {
		return pve.ClientNew{}, err
	}
	return c.New(), nil
}

func GetFlagDescription() string {
	v, _ := RootCmd.Flags().GetString(flagDescription)
	return v
}

func FlagDryRun(cmd *cobra.Command, p *bool) {
	cmd.Flags().BoolVar(p, flagDryRun, false, "logs operations that would be executed without making changes")
}

func FlagKeep(cmd *cobra.Command, p *int, d int) {
	cmd.Flags().IntVarP(p, flagKeep, "k", d, "number of snapshot to keep after pruning")
}

func GetFlagLabel() string {
	v, _ := RootCmd.Flags().GetString(flagLabel)
	return v
}

func GetFlagCron() *string {
	if !RootCmd.Flags().Changed(cron.Flag) {
		return nil
	}
	v, _ := RootCmd.Flags().GetString(cron.Flag)
	return &v
}

func GetFlagDatetime() (string, error) {
	v, _ := RootCmd.Flags().GetString(datetime.Flag)
	return datetime.Parse(v)
}

func GetFlagMaxParallel() uint {
	v, _ := RootCmd.Flags().GetUint(flagMaxParallel)
	return v
}

func GetFlagMaxParallelNode() *uint {
	if !RootCmd.Flags().Changed(flagMaxParallelNode) {
		return nil
	}
	v, _ := RootCmd.Flags().GetUint(flagMaxParallelNode)
	return &v
}

func GetFlagOutput(cmd *cobra.Command) (*os.File, error) {
	if !cmd.Flags().Changed(FlagOutput) {
		return os.Stdout, nil
	}
	path, _ := cmd.Flags().GetString(FlagOutput)
	if path == "" {
		return nil, errors.New("--" + FlagOutput + " may not be empty")
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}

func SetFlagOutput(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagOutput, "o", "", "output location, stdout when unset")
}

type Shared struct {
	Client      pve.ClientNew
	Cron        *string
	TimeFormat  string
	Ctx         context.Context
	Description string
	Filter      *filter.Filter
	Format      format.Enum
	Label       string
	Pool        pool.Pool
}

func Setup() (shared Shared, err error) {
	shared = Shared{
		Cron:        GetFlagCron(),
		Ctx:         RootCmd.Context(),
		Description: GetFlagDescription(),
		Format:      outputFormat,
		Label:       GetFlagLabel(),
	}
	if shared.TimeFormat, err = GetFlagDatetime(); err != nil {
		return
	}
	if shared.Filter, shared.Client, err = NewFilterAndClient(shared.Ctx); err != nil {
		return
	}

	nodePtr := GetFlagMaxParallelNode()
	var nodes []pve.NodeName
	var nodeP uint
	if nodePtr != nil {
		if nodes, err = proxmox.ListNodes(shared.Ctx, shared.Client.Node); err != nil {
			return
		}
		nodeP = *nodePtr
	}
	shared.Pool = pool.New(GetFlagMaxParallel(), nodes, nodeP)
	return
}
