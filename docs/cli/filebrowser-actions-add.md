# filebrowser actions add

Define an action, replacing any with the same name

## Synopsis

Define an action, replacing any with the same name.

The command is a program, given by its absolute path, followed by fixed
arguments. It is started directly, without a shell. One argument may be
$FILES: it is replaced by the paths the user passes, each resolved to a
file inside that user's scope. Without $FILES the user cannot pass
anything. Put "--" before $FILES when the program supports it, and quote
the command so your own shell does not expand it:

  filebrowser actions add ls '/bin/ls -la -- $FILES'

```
filebrowser actions add <name> <command> [flags]
```

## Options

```
  -h, --help[=true|false]   help for add
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser actions](filebrowser-actions.md)	 - Command runner actions management utility

