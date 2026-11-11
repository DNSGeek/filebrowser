# filebrowser cmds add

Add a command to run on a specific event

## Synopsis

Add a command to run on a specific event, such as before_upload or
after_delete.

The command is started directly, without a shell: there are no pipes,
redirections or command substitutions. An argument that is exactly $FILE,
$SCOPE, $TRIGGER, $USERNAME or $DESTINATION is replaced by its value, and
the same variables are set in the command's environment. Quote the command
so your own shell does not expand them first:

  filebrowser cmds add after_upload '/usr/local/bin/scan $FILE'

End the command with & to run it without waiting for it to finish.

```
filebrowser cmds add <event> <command> [flags]
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

* [filebrowser cmds](filebrowser-cmds.md)	 - Event hooks management utility

