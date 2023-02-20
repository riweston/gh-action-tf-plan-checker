# Overview

This is to provide a simple way to validate if a terraform output plan performs actions which are explicitly allowed
using a CLI tool and return a non-zero exit code if that it is the case.

## Usage

```bash
$ terraform plan -out=plan.out
...
  # local_file.this must be replaced
-/+ resource "local_file" "this" {
      ~ directory_permission = "0777" -> "0775" # forces replacement
      ~ id                   = "0a0a9f2a6772942557ab5355d76af442f8f65e01" -> (known after apply)
        # (3 unchanged attributes hidden)
    }

Plan: 1 to add, 0 to change, 1 to destroy.
...

$ tfplan-checker --tfplan plan.out
2023/02/20 08:55:29 Deny Changes: [delete create]

$ echo $?
1

```

## Flags

```bash
$ tfplan-checker --help
NAME:
   tfplan-checker - A new cli application

USAGE:
   tfplan-checker [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --tfplan FILE, -p FILE  Load terraform plan from FILE
   --allow-delete, -d      Allow delete actions (default: false)
   --allow-update, -u      Allow update actions (default: false)
   --allow-create, -c      Allow create actions (default: false)
   --help, -h              show help
```
