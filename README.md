ssm-run-command
===============

Invoke a remote command(s) using SSM RunCommand and stream results back to stderr/stdout

```text
Usage:
  run-command [flags]

Flags:
  -c, --comment string           (Optional) Comment for command visible 
                                 on the SSM dashboard (default "invoked using ssm-run-command CLI")
      --execution-timeout int    (Optional) The time in seconds for a command 
                                 to complete before it is considered to
                                 have failed. Default is 3600 (1 hour). Maximum is 172800 (48 hours). (default 3600)
  -h, --help                     help for run-command
  -l, --log-group string         (Optional) The AWS CloudWatch log group 
                                 for RunCommand to log to (default "/ssm-run-command")
      --max-concurrency string   (Optional) The maximum number of instances that 
                                 are allowed to run the command at the same time. You can 
                                 specify a number such as 10 or a percentage such as 10%. (default "50")
      --max-errors string        (Optional) The maximum number of errors allowed without 
                                 the command failing. When the command fails one more time beyond the value 
                                 of MaxErrors, the systems stopnsending the command to additional targets. 
                                 You can specify a number like 10 or a percentage like 10%. (default "1")
      --target stringArray       Target instances with these values. 
                                 For example: --target tag:App=MyApplication --target tag:Environment=qa
  -t, --target-limit int         (Optional) Limit execution to first n targets (default 50)
      --version                  version for run-command
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
- if command is cancelled, exit 1
- delete logs from cloudwatch before closing (add --keep-logs flag to prevent deleting them from cloudwatch)
- if no targets are found after n seconds, exit 1 (--init-timeout)
- provide documentation around the required AWS IAM permissions used
