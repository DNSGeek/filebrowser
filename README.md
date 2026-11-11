> [!NOTE]
>
> **This is a maintained fork of the unmaintained [`filebrowser/filebrowser`](https://github.com/filebrowser/filebrowser).** Version 3 changes the Go module path to `github.com/DNSGeek/filebrowser/v3` and receives security fixes.

<p align="center">
  <img src="./branding/banner.png" width="550"/>
</p>

File Browser provides a file managing interface within a specified directory and it can be used to upload, delete, preview and edit your files. It is a **create-your-own-cloud**-kind of software where you can just install it on your server, direct it to a path and access your files through a nice web interface.

**Background:** upstream [announced](https://hacdias.com/2026/07/28/filebrowser/) in July 2026 that it would no longer be maintained.

## Security

Published advisories are listed under [security advisories](https://github.com/DNSGeek/filebrowser/security/advisories),
and reporting instructions are in [SECURITY.md](SECURITY.md). One known issue class
remains only partly addressed:

- **Interactive command runner.** Redesigned: users can only start actions the operator defined from the command line on the host, with fixed arguments, and can pass nothing but files inside their own scope. No shell is involved. What a program does once started is not sandboxed, so an action such as `git` or `find` is still equivalent to shell access; it therefore stays disabled by default (`--disable-exec=false` enables it). See [`docs/command-execution.md`](docs/command-execution.md). Background: [#5199](https://github.com/filebrowser/filebrowser/issues/5199).
- **Event hooks.** Redesigned: hooks are defined only from the command line on the host, never from the web interface, and run without a shell. They are enabled by default, with none defined. Hooks stored by earlier versions are not run until re-added. See [`docs/command-execution.md`](docs/command-execution.md).

Hardening guidance:

- **Do not expose it directly to the internet.** Put it behind a reverse proxy that terminates TLS and performs its own authentication.
- **Keep the interactive command runner disabled unless you need it.** It is off by default. If you enable it, define only narrow actions. See [#5199](https://github.com/filebrowser/filebrowser/issues/5199) and [`docs/command-execution.md`](docs/command-execution.md).
- **Run it unprivileged, inside a container**, with only the directory you intend to serve mounted into it.

## Documentation

Documentation on how to install, configure, and build this project lives in [`docs`](docs) in this repository.

[CONTRIBUTING.md](CONTRIBUTING.md) documents how to build and develop the project, which remains useful to anyone forking it.

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
