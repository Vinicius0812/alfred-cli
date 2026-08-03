# Security Policy

Alfred is designed for explicit, auditable development automation. Security
reports are welcome, especially when they involve configuration parsing, path
validation, external command execution, secret handling, installation, or the
release supply chain.

## Supported Versions

Alfred is currently an early preview. Security fixes are provided for the
latest published release line and the current `master` branch.

| Version | Supported |
| --- | --- |
| `master` | Yes |
| `0.1.x` | Yes, until `0.2.0` is released |
| `< 0.1.0` | No |

## Reporting a Vulnerability

Please use GitHub's private vulnerability reporting feature in the repository
Security tab when it is available. Include:

- the affected Alfred version and operating system;
- a minimal reproduction or proof of concept;
- the expected and observed behavior;
- the potential impact;
- any suggested mitigation.

Do not include credentials, tokens, private source code, or personal data in a
public issue. If private reporting is unavailable, open a public issue asking
the maintainer for a private contact channel without disclosing exploit details.

Please allow the maintainer time to reproduce and remediate the issue before
public disclosure. Reports will be acknowledged and assessed as promptly as
the volunteer maintenance schedule allows; no fixed response SLA is promised
during the preview phase.

## Security Boundaries

- Project configuration is treated as code. Review `alfred.yaml` and every
  locally extended preset before using commands that execute workflows.
- Alfred validates configuration but is not an operating-system sandbox.
- Secrets should be supplied through the environment and must not be committed
  to project configuration.
- Alfred must not download remote presets or execute a workflow without an
  explicit user action.
- Destructive, networked, or Git-mutating operations should be inspectable and
  require the configured confirmation policy.

The current `v0.1.x` implementation validates configuration and the local
environment but does not yet execute configured workflows.

## In Scope

- YAML ambiguity, parser denial of service, or configuration bypasses;
- path traversal and symbolic-link escapes;
- command or argument injection;
- accidental secret disclosure in output or logs;
- checksum verification bypasses;
- compromised or non-reproducible release behavior;
- privilege escalation caused by Alfred.

Requests to add offensive execution, persistence, credential theft, evasion,
or destructive automation are outside the project's intended scope.
