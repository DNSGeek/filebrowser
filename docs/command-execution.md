# Command Execution

File Browser can run commands on the host in two separate ways. They have different risks and are enabled separately.

| Feature           | Default  | Flag                   | Who decides what runs                               |
| ----------------- | -------- | ---------------------- | --------------------------------------------------- |
| Event hooks       | enabled  | `--disable-hooks`      | Only whoever can run `filebrowser cmds` on the host |
| Interactive shell | disabled | `--disable-exec=false` | Only whoever can run `filebrowser actions` on host  |

## Event Hooks

A hook is a command that runs before or after a file event. The events are:

- Copy
- Rename
- Upload
- Delete
- Save

No hooks are defined by default, so nothing runs until you add one.

### Managing hooks

Hooks are managed only from the command line, on the machine File Browser runs on:

```bash
filebrowser cmds add after_upload '/usr/local/bin/scan $FILE'
filebrowser cmds rm after_upload 0
filebrowser cmds ls
```

They cannot be read or changed from the web interface or the HTTP API. Defining what runs on the host therefore takes access to the host, and a compromised administrator account cannot be used to add a hook.

Stop File Browser before running `filebrowser cmds`, as the database can only be opened by one process.

### How hooks run

A hook is split into a program and its arguments and started directly. It never goes through a shell, so pipes, redirections, `;`, `&&` and command substitutions are not available. There is no shell setting any more.

These variables describe the event:

- `FILE` with the full absolute path to the changed file.
- `SCOPE` with the path to user's scope.
- `TRIGGER` with the name of the event, such as `after_upload`.
- `USERNAME` with the user's username.
- `DESTINATION` with the absolute path to the destination. Only used for **copy** and **rename.**

They are available in two ways:

- **As a whole argument.** An argument that is exactly `$FILE` or `${FILE}` (and likewise for the others) is replaced by the value. It reaches the program as a single argument, whatever characters it contains.
- **In the environment.** The same variables are set in the command's environment.

A variable inside a longer argument, such as `--out=$FILE` or `$FILE.bak`, is **not** replaced. File names and user names are chosen by users, so they are never pasted into the middle of an argument.

If a value starts with `-`, the hook fails instead of passing it as an argument, because the program would read it as an option. Before-hooks that fail stop the action.

End a hook with `&` to run it without waiting for it to finish.

### Hooks that need a shell

If you need shell features, call the shell yourself and read the variables from the environment, always quoted:

```bash
filebrowser cmds add after_upload 'sh -c '\''printf "%s\n" "$FILE" >> /var/log/uploads.log'\'''
```

This is safe because `$FILE` inside the script is expanded by the shell from the environment, as data. Do not build a script by other means from these values. A script file is usually simpler:

```bash
filebrowser cmds add after_upload '/usr/local/bin/on-upload.sh'
```

### Upgrading from earlier versions

Earlier versions stored hooks in a setting that could be edited from the web interface, and ran them through the configured shell when the command runner was enabled. Those stored hooks are **no longer run**. File Browser logs a warning at startup while any remain, and `filebrowser cmds ls` lists them under a separate heading.

Review them, then add the ones you want again with `filebrowser cmds add`, adjusting any that relied on a shell or on variables inside a longer argument.

## Interactive Shell

> [!CAUTION]
>
> The interactive shell is disabled by default. It lets users start programs on the host, and File Browser controls only which program runs and which files it is given, not what the program then does. Read [Choosing safe actions](#choosing-safe-actions) before enabling it. Background: issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199).

Within File Browser you can toggle the shell (`< >` icon at the top right) and this will open a command window at the bottom of the screen. This functionality can be turned on using the environment variable `FB_DISABLE_EXEC=false` or the flag `--disable-exec=false`.

Despite the name, it is not a shell. A user can only start an **action**: a command the operator defined in advance, by name, followed by the files to run it on.

### Defining actions

Actions are managed only from the command line, on the machine File Browser runs on:

```bash
filebrowser actions add ls '/bin/ls -la -- $FILES'
filebrowser actions add unzip '/usr/bin/unzip -n -- $FILES'
filebrowser actions rm unzip
filebrowser actions ls
```

Like hooks, they cannot be read or changed from the web interface or the HTTP API. Stop File Browser before running `filebrowser actions`.

An action is a program followed by fixed arguments:

- The program must be given by its **absolute path**.
- It is started directly, never through a shell. The `--shell` option of earlier versions is ignored.
- One argument may be `$FILES`. It is replaced by the files the user names. Without it, the user cannot pass anything.
- Put `--` before `$FILES` whenever the program supports it.

### Allowing users to run actions

A user can run an action when its name is in their command list: Settings > User Management > (edit user) > Commands, or Settings > Global Settings for the default applied to new users. The user also needs the **Execute** permission.

A name in a user's command list that matches no action does nothing, so command lists from earlier versions grant nothing until you define actions with those names.

### What a user can type

The first word is the action's name. Every other word is taken as a file name, relative to the directory the user is browsing (or to the root of their scope if it starts with `/`):

```
ls
ls photos "holiday 2026"
unzip archive.zip
```

Each name is passed to the program as the full path of that file on disk, and only if the user may reach it: it must be inside their scope, must not lead out of it through a symbolic link (unless `--followExternalSymlinks` is set), and must not be denied by a rule. Otherwise the command is refused.

This means a user cannot add options (`ls -R` looks for a file named `-R`), cannot name files outside their scope, and cannot use shell syntax, wildcards or pipes.

The program runs in the directory the user is browsing, with a minimal environment that excludes File Browser's own settings and secrets, for at most ten minutes.

### Choosing safe actions

File Browser decides which program starts and which paths it receives. It does not sandbox the program. Before defining an action, consider what the program does with files the user controls:

- **Programs that run other programs or load code from the files they process are equivalent to giving the user a shell.** Examples: `git` (hooks and configuration inside the repository), `make`, `find` with `-exec`, interpreters such as `sh`, `python` or `node`, and editors.
- **Programs that descend into directories** are not checked against rules or symbolic links below the paths the user named. `ls -R`, `tar` or `cp -r` on a directory will reach files that a rule hides from that user, and may follow symbolic links that point outside the scope.
- **Programs that extract archives** can create symbolic links and, depending on the tool and its options, write outside the target directory.
- Every action runs as the user File Browser runs as, with that user's access to the whole host.

Prefer narrow actions with fixed options, run File Browser unprivileged in a container that mounts only the served directory, and leave the feature disabled if you do not need it.

> [!NOTE]
>
> If using a proxy manager then remember to enable websockets support for the File Browser proxy

> [!NOTE]
>
> If using Docker and you want to add a new command that is not in the base image then you will need to build a custom Docker image using `ghcr.io/dnsgeek/filebrowser` as a base image. For example to add 7z:
>
> ```docker
> FROM ghcr.io/dnsgeek/filebrowser
> RUN sudo apt install p7zip-full
> ```
