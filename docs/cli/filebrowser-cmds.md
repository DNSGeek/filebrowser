# filebrowser cmds

Event hooks management utility

## Synopsis

Event hooks management utility.

Hooks are commands that run before or after a file event (save, copy,
rename, upload, delete). They can only be managed with this command, on
the machine File Browser runs on, and are started directly, without a
shell. See docs/command-execution.md.

## Options

```
  -h, --help[=true|false]   help for cmds
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser](filebrowser.md)	 - A stylish web-based file browser
* [filebrowser cmds add](filebrowser-cmds-add.md)	 - Add a command to run on a specific event
* [filebrowser cmds ls](filebrowser-cmds-ls.md)	 - List all commands for each event
* [filebrowser cmds rm](filebrowser-cmds-rm.md)	 - Removes a command from an event hooker

