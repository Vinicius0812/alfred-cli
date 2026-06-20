# Alfred CLI

Alfred is an open source developer assistant for automating repetitive project
tasks with predictable, versioned configuration.

The first release focuses on:

- validating local development environments;
- creating standardized Conventional Commits;
- managing Docker Compose projects;
- running project workflows declared in YAML;
- allowing teams and companies to share reusable presets.

Alfred's core stays generic. Company-specific branch rules, commit policies,
Docker settings, and workflows belong in configuration, not in the executable.

## Planned Commands

```text
alfred init
alfred doctor
alfred commit
alfred docker up
alfred docker down
alfred run <workflow>
alfred config validate
```

## Configuration

Projects are configured through `alfred.yaml`:

```yaml
version: 1

project:
  name: example-api

commit:
  convention: conventional
  confirm_push: true

docker:
  compose_files:
    - compose.yaml

workflows:
  test:
    description: Run the project test suite
    steps:
      - run: go test ./...
```

See [Product Scope](docs/product.md), [Configuration v1](docs/configuration-v1.md),
and the [examples](examples/) for the current design.

## Status

Alfred is in its specification phase. The configuration contract is being
defined before implementation to keep the CLI portable and avoid coupling it
to the internal rules of its predecessor, UNI CLI.

## License

A license has not been selected yet. An OSI-approved permissive license should
be chosen before the first public release.
