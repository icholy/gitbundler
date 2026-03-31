# gitbundler

A server that maintains up-to-date [git bundles](https://git-scm.com/docs/git-bundle) for a set of repositories and serves them over HTTP.

## Usage

Create a `config.yaml`:

```yaml
data_dir: data
addr: ":8080"
repos:
  - name: myrepo
    url: https://github.com/owner/repo.git
    interval: 5m
```

Run the server:

```
go run . -config config.yaml
```

Fetch a bundle:

```
curl -o myrepo.bundle http://localhost:8080/myrepo.bundle
git clone myrepo.bundle myrepo
```

## Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `data_dir` | Directory for clones and bundles | `data` |
| `addr` | HTTP listen address | `:8080` |
| `repos[].name` | Name used in the bundle URL | required |
| `repos[].url` | Git remote URL | required |
| `repos[].interval` | How often to fetch and re-bundle | `5m` |

### Environment Variables

Use `${env:NAME}` in the config to reference environment variables. This is useful for private repos:

```yaml
repos:
  - name: private-repo
    url: https://x-access-token:${env:GITHUB_TOKEN}@github.com/org/private-repo.git
```

## Docker

```
docker compose up --build
```
