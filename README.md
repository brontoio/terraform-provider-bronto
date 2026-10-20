# Terraform Provider for Bronto

This is the official Terraform provider for [Bronto](https://bronto.io), letting you
manage Bronto resources - users, dashboards, widgets, and metric definitions - as
code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5
- [Go](https://go.dev/doc/install) >= 1.23 (only required if building the provider)

## Using the provider

```hcl
terraform {
  required_providers {
    bronto = {
      source  = "brontoio/bronto"
      version = "~> 0.1"
    }
  }
}

provider "bronto" {
  # endpoint and api_key can also be supplied via the BRONTO_ENDPOINT and
  # BRONTO_API_KEY environment variables.
  endpoint = "https://api.eu.bronto.io"
  api_key  = var.bronto_api_key
}

resource "bronto_user" "example" {
  first_name = "Jane"
  last_name  = "Doe"
  email      = "jane.doe@example.com"
  roles      = ["Standard"]
}
```

Full documentation, including every resource and data source, is published on the
[Terraform Registry](https://registry.terraform.io/providers/brontoio/bronto/latest/docs)
and mirrored in this repository under [`docs/`](./docs).

### Resources and data sources

| Name | Description |
| --- | --- |
| `bronto_user` | Manages a Bronto user. |
| `bronto_dashboard` | Manages a Bronto dashboard. |
| `bronto_widget` | Manages a Bronto widget. |
| `bronto_metric_definition` | Manages a Bronto metric definition. |
| `data.bronto_user` | Looks up an existing user by ID or email. |

Tags are supported on `bronto_user`, `bronto_dashboard`, and `bronto_widget`.
`bronto_metric_definition` does not support tags - the underlying API doesn't
expose them for that resource type.

Multiple orgs (e.g. a root org and a child org, each with their own API key) are
supported by declaring multiple aliased instances of the provider and pointing
each resource at the right one with `provider = bronto.<alias>`; see
[`docs/index.md`](./docs/index.md).

## Developing the provider

```shell
go install .                 # build and install to $GOPATH/bin
go build ./...               # build only
go vet ./...
go test ./...
```

To try a local build against real Terraform configuration without publishing
anything, point Terraform at your locally-built binary with a `dev_overrides`
block in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "brontoio/bronto" = "/path/to/your/GOBIN"
  }
  direct {}
}
```

### Regenerating docs

Documentation under `docs/` is generated from the resource schemas and the
example `.tf` files under `examples/` with
[`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs). After
changing a schema or an example, regenerate it with:

```shell
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs generate --provider-name bronto
```

(`--provider-name bronto` is required because this repository's directory name
doesn't follow the `terraform-provider-<name>` convention `tfplugindocs`
assumes by default.)

### Running tests

```shell
go test ./...
```

Tests are unit tests against the provider's internal encode/decode and
reconciliation logic; they don't require network access or a live Bronto API
key.

## License

[Mozilla Public License 2.0](./LICENSE).
