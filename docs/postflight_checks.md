# Postflight checks
Postflight checks are performed after qliksense is installed on the cluster and during normal operating mode of the product. Such checks can range from validating certain conditions to checking the status of certain operations or entities on the kubernetes cluster.

Run the following command to view help about the commands supported by postflight at any moment:
```
$ qliksense postflight
perform postflight checks on the cluster

Usage:
  qliksense postflight [command]

Examples:
qliksense postflight <postflight_check_to_run>

Available Commands:
  db-migration-check check mongodb migration status on the cluster

Flags:
  -h, --help      help for postflight
  -v, --verbose   verbose mode
```

### Run all postflight checks
This command runs all the postflight checks available.

```shell
$ qliksense postflight all  
Running all postflight checks...

Postflight db migration check... 
Logs from pod: qliksense-users-6977cb7788-qlgmv
{"caller":"main.go:39","environment":"qseok","error":"error parsing uri: scheme must be \"mongodb\" or \"mongodb+srv\"","level":"error","message":"failed to connect to ","timestamp":"2020-06-17T04:10:11.7891913Z","version":""}
To view more logs in this context, please run the command: kubectl logs -n test_ns qliksense-users-6977cb7788-qlgmv migration
PASSED

All postflight checks have PASSED
```

### DB migration check
This command checks init containers for successful database migrarion completions, and reports failure, if any to the user.

An example run of this check produces an output as shown below:

```shell
$ qliksense postflight db-migration-check   
Logs from pod: qliksense-users-6977cb7788-cxxwh
{"caller":"main.go:39","environment":"qseok","error":"error parsing uri: scheme must be \"mongodb\" or \"mongodb+srv\"","level":"error","message":"failed to connect to ","timestamp":"2020-06-01T01:07:18.4170507Z","version":""}
To view more logs in this context, please run the command: kubectl logs -n test_ns qliksense-users-6977cb7788-qlgmv migration
PASSED
Postflight db_migration_check completed
```