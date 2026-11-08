# filebrowser users add

Create a new user

## Synopsis

Create a new user and add it to the database.

```
filebrowser users add <username> <password> [flags]
```

## Options

```
      --aceEditorTheme string                ace editor's syntax highlighting theme for users
      --commands strings                     a list of the commands a user can execute
      --dateFormat[=true|false]              use date format (true for absolute time, false for relative)
  -h, --help[=true|false]                    help for add
      --hideDotfiles[=true|false]            hide dotfiles in file listings
      --locale string                        locale for users (default "en")
      --lockPassword[=true|false]            lock password
      --perm.admin[=true|false]              admin perm for users
      --perm.create[=true|false]             create perm for users (default true)
      --perm.delete[=true|false]             delete perm for users (default true)
      --perm.download[=true|false]           download perm for users (default true)
      --perm.execute[=true|false]            execute perm for users (default true)
      --perm.modify[=true|false]             modify perm for users (default true)
      --perm.rename[=true|false]             rename perm for users (default true)
      --perm.share[=true|false]              share perm for users (default true)
      --redirectAfterCopyMove[=true|false]   redirect to destination after copy/move
      --scope string                         scope for users (default ".")
      --singleClick[=true|false]             use single clicks only
      --sorting.asc[=true|false]             sorting by ascending order
      --sorting.by string                    sorting mode (name, size or modified) (default "name")
      --viewMode string                      view mode for users (default "list")
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [filebrowser users](filebrowser-users.md)	 - Users management utility

