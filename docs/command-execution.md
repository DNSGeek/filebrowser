# Command Execution

File Browser can run commands on the host in two separate ways. They have different risks and are enabled separately.

| Feature           | Default  | Flag                   | Who decides what runs                               |
| ----------------- | -------- | ---------------------- | --------------------------------------------------- |
| Event hooks       | enabled  | `--disable-hooks`      | Only whoever can run `filebrowser cmds` on the host |
| Interactive shell | disabled | `--disable-exec=false` | Administrators, in the web interface                |

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

A hook is split into a program and its arguments and started directly. It never goes through a shell, so pipes, redirections, `;`, `&&` and command substitutions are not available. The **shell** setting does not apply to hooks.

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
> The interactive shell is disabled by default due to continuous and known security vulnerabilities. A user who may run a command can read and change anything the File Browser process can, regardless of their scope. Only enable it if you are aware of the risks. For more information, consult issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199).

Within File Browser you can toggle the shell (`< >` icon at the top right) and this will open a shell command window at the bottom of the screen. This functionality can be turned on using the environment variable `FB_DISABLE_EXEC=false` or the flag `--disable-exec=false`.

By default no commands are available as the command list is empty. To enable commands these need to either be done on a per-user basis (including for the Admin user).

You can do this by adding them in Settings > User Management > (edit user) > Commands or to _apply to all new users created from that point forward_ they can be set in Settings > Global Settings

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
