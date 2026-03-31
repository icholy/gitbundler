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
  - name: private-repo
    url: https://x-access-token:${env:GITHUB_TOKEN}@github.com/org/private-repo.git
    interval: 10m
```

Add it to your `docker-compose.yml`:

```yaml
services:
  gitbundler:
    image: ghcr.io/icholy/gitbundler:latest
    ports:
      - "8080:8080"
    working_dir: /app
    environment:
      - GITHUB_TOKEN
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - data:/app/data

volumes:
  data:
```

Clone using the bundle for the initial download:

```
git clone --bundle-uri=http://localhost:8080/myrepo.bundle https://github.com/owner/repo.git
```

## Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `data_dir` | Directory for clones and bundles | `data` |
| `addr` | HTTP listen address | `:8080` |
| `repos[].name` | Name used in the bundle URL | required |
| `repos[].url` | Git remote URL | required |
| `repos[].interval` | How often to fetch and re-bundle | `5m` |

Environment variables can be referenced in the config using `${env:NAME}` syntax.
