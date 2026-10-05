# Security Policy

## Supported Versions

Security fixes are released for the latest `3.x` version only.

| Version | Supported |
| ------- | --------- |
| 3.x     | ✅        |
| 2.x     | ❌        |
| < 2.0   | ❌        |

Version 3 is a maintained fork of the unmaintained upstream project
(`filebrowser/filebrowser`, last release 2.x). Upstream advisories that apply to
2.x are not tracked here unless the affected code is still present in 3.x.

## Before Reporting

Please check the [existing advisories](https://github.com/DNSGeek/filebrowser/security/advisories) and open issues first, and confirm:

- **It concerns this fork's code at its latest release.** Reports about code, features, or endpoints that don't exist here belong to the relevant project.
- **It isn't an already-known class** that remains unaddressed. Those are listed under [Security](README.md#security) in the README; reports covering them are likely to be closed as duplicates.

## Reporting a Vulnerability

Report privately via the [Security](https://github.com/DNSGeek/filebrowser/security) page.

Please include, where possible:

- The commit or version the issue was found at
- A plaintext proof of concept (no binaries)
- Steps to reproduce
- Recommended remediation, if any

Confirmed issues are fixed in the next `3.x` release and then published as advisories.
