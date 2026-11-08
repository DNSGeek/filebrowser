# filebrowser rules add

Add a global rule or user rule

## Synopsis

Add a global rule or user rule.

```
filebrowser rules add <path|expression> [flags]
```

## Options

```
  -a, --allow[=true|false]   indicates this is an allow rule
  -h, --help[=true|false]    help for add
  -r, --regex[=true|false]   indicates this is a regex rule
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
  -i, --id uint           id of user to which the rules apply
  -u, --username string   username of user to which the rules apply
```

## See Also

* [filebrowser rules](filebrowser-rules.md)	 - Rules management utility

