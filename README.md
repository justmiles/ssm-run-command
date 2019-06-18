ssm-run-command
===============

Invoke a remote command(s) using SSM RunCommand and stream results back to stderr/stdout

```text
Usage:
  run-command [flags] <command>

Flags:
  -h, --help                 help for run-command
      --target stringArray   target instances with these values. Example: --target "tag:App=MyApplication" --target "tag:Environment=qa"
      --target-limit int     limit execution to first n targets. Max 50 (default 50)
      --version              version for run-command
```

Examples:

```bash
# list files on EC2 instance named "MyApplication"
ssm-run-command --target "tag:Name=MyApplication" ls -al
```

```bash
# Show root disk usage on all instances with a tagged called "Environment" set to "qa"
ssm-run-command --target "tag:Environment=qa" df -h /

--> 2019/06/09 11:44:18 stdout i-xxxxxxxxxxxxxxxx1 - 1560444254128 Filesystem      Size  Used Avail Use% Mounted on
--> 2019/06/09 11:44:18 stdout i-xxxxxxxxxxxxxxxx1 - 1560444254128 /dev/xvda1       32G  6.2G   26G  20% /
--> 2019/06/09 11:44:29 stdout i-xxxxxxxxxxxxxxxx2 - 1560444264995 Filesystem      Size  Used Avail Use% Mounted on
--> 2019/06/09 11:44:29 stdout i-xxxxxxxxxxxxxxxx2 - 1560444264995 /dev/xvda1       32G  9.9G   23G  31% /
--> 2019/06/09 11:44:39 stdout i-xxxxxxxxxxxxxxxx3 - 1560444275921 Filesystem      Size  Used Avail Use% Mounted on
--> 2019/06/09 11:44:39 stdout i-xxxxxxxxxxxxxxxx3 - 1560444275921 /dev/xvda1      246G   57G  177G  25% /
```

TODO:

- capture ctrl+c and cancel the command
- set custom cloudwatch logs location
- delete logs from cloudwatch before closing (add --keep-logs flag to prevent deleting them from cloudwatch)
- if command is cancelled, exit 1
- if no targets are found after n seconds, exit 1 (--init-timeout)
- provide documentation around the required AWS IAM permissions used
