# go-pve-autosnap

![Coverage](https://img.shields.io/badge/Coverage-76.1%25-brightgreen)

With `go-pve-autosnap` you can manage your [PVE](https://www.proxmox.com/en/products/proxmox-virtual-environment/overview) snapshots.  
Do one-time operations or schedule automatic snapshots.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Quick Start](#quick-start)
- [Options](#options)
- [Filtering](#filtering)
  - [Filtering Examples](#filtering-examples)

## Quick Start

Create snapshots hourly and prune according to the retention limit:

```bash
go-pve-autosnap --url=https://127.0.0.1:8006/api2/json --username=root@pam --password-file=/path/to/password --filter=@all: --keep=12 --cron="0 * * * *" snap
```

Remove snapshots exceeding the configured retention limit:

```bash
go-pve-autosnap --url=https://127.0.0.1:8006/api2/json --username=root@pam --password-file=/path/to/password --filter=@all: --keep=0 prune
```

Summarize all guests and snapshots that match the filter, label and description.

```bash
go-pve-autosnap --url=https://127.0.0.1:8006/api2/json --username=root@pam --password-file=/path/to/password --filter=@all: --datetime=YY-MM-DD_hh-mm-ss --label=summary summary
```

## Options

| Short Flag| Long Flag            | Description
|:----------|:---------------------|:-----------
| `-u`      | `--url`              | PVE API URL, for example: `https://127.0.0.1:8006/api2/json`.
| `-i`      | `--insecure`         | Ignores HTTPS certificate errors.
| `-n`      | `--username`         | Inline PVE username, for example: `root@pam`.
| `-N`      | `--username-file`    | Path to the file containing the PVE username.
| `-p`      | `--password`         | Inline PVE password.
| `-t`      | `--token`            | Inline PVE API token.
| `-T`      | `--token-file`       | Path to the file containing the PVE API token.
| `-P`      | `--password-file`    | Path to the file containing the PVE password.
|           | `--dry-run`          | When set, no changes are made. All operations are logged as if they were executed.
| `-f`      | `--filter`           | Filter for selecting guests, see [Filtering](#filtering), defaults to `@all:`.
| `-d`      | `--description`      | PVE snapshot description.
| `-l`      | `--label`            | Prefix of the snapshot name, this will be followed by a timestamp. Defaults to `auto-`.
| `-k`      | `--keep`             | Number of snapshots to keep after pruning. Only snapshots with the configured description and label are eligible for pruning. To ignore pruning set to `-1`.
| `-c`      | `--cron`             | When `--cron` is set, the application runs continuously in daemon mode and executes snapshots according to the supplied cron schedule. A scheduled snapshot job is skipped if the previous job is still running.
|           | `--max-parallel`     | Maximum number of guests processed in parallel, defaults to `1`.
|           | `--max-parallel-node`| Maximum number of guests processed in parallel on a single node.
|           | `--format`           | Output format, options are `cli`, `json`, `json-pretty`. Defaults to `cli`.
|           | `--timeout`          | API timeout in seconds. Defaults to `5`.
|           | `--datetime`         | Date and time format, defaults to `YYYYMMDDhhmmss`.
| `-s`      | `--state`            | Whether memory state should be included in snapshots, defaults to `false`.
| `-o`      | `--output`           | Output file, defaults to stdout.

## Filtering

The `--filter` parameter is used to include and exclude guest systems.
A filter starts with an `@` followed by its optional inclusion modifier, `+` for inclusion, `-` for exclusion.
When no modifier is provided it defaults to `+`.
Most filters can take multiple values delimited by `,`, `;`, or space.
The filter is applied in steps from left to right.  

Example:  
`--filter="@all: @-pool:demo @+id:100,300-500 @-node:pve2;pve3 @-tag:no_snapshot"`

| Pattern    | Example                     | Description
|:-----------|:----------------------------|:-----------
| All        | `--filter=@all:`            | The `@all:` filter selects all guests and is typically used as the first filter.
| ID Single  | `--filter=@id:100`          | Select a guest by ID.
| ID Range   | `--filter=@id:100-999`      | Select guests whose ID is within the range.
| Name Single| `--filter=@name:test-server`| Select a guest by name.
| Tag Single | `--filter=@tag:demo`        | Select guests by tag.
| Node Single| `--filter=@node:pve05`      | Select guests based on the node they are running on.
| Pool Single| `--filter=@pool:production` | Select guests based on pool membership.
| Exclude    | `--filter=@-id:100`         | By adding a `-` between the `@` and filter name it will subtract instead of add.

### Filtering Examples

Select all guests except those in the `demo` pool:  
`--filter="@all: @-pool:demo"`

Select only guests with IDs `100` through `200`:  
`--filter=@id:100-200`

Select all guests tagged as `production` except when tagged as `no_snapshot`:  
`--filter='@tag:production @-tag:no_snapshot'`

Select all guests except those in the `demo` pool that don't have the `important` tag:  
`--filter="@all: @-pool:demo @+tag:important"`
