# filebrowser actions

Command runner actions management utility

## Synopsis

Command runner actions management utility.

An action is a command users can start by name from the command runner in
the web interface, when it is enabled with --disable-exec=false. Actions
can only be managed with this command, on the machine File Browser runs
on. A user can run the actions listed in their "commands" setting. See
docs/command-execution.md.

## Options

```
  -h, --help[=true|false]   help for actions
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser](filebrowser.md)	 - A stylish web-based file browser
* [filebrowser actions add](filebrowser-actions-add.md)	 - Define an action, replacing any with the same name
* [filebrowser actions ls](filebrowser-actions-ls.md)	 - List all actions
* [filebrowser actions rm](filebrowser-actions-rm.md)	 - Remove an action

