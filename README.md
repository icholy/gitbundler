# gitbundler

A server that maintains up-to-date [git bundles](https://git-scm.com/docs/git-bundle) for a set of repositories and serves them over HTTP. Use it with `git clone --bundle-uri` to speed up cloning large repositories by bootstrapping from a pre-built bundle instead of fetching all objects from the remote.

## Usage

Create a `gitbundler.yaml`:

```yaml
repos:
  - name: myrepo
    url: https://x-access-token:${env:GITHUB_TOKEN}@github.com/owner/repo.git
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
      GITHUB_TOKEN: ${GITHUB_TOKEN}
    volumes:
      - ./gitbundler.yaml:/app/gitbundler.yaml:ro
      - gitbundler-data:/app/data

volumes:
  gitbundler-data:
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
| `max_concurrent` | Max concurrent clone/fetch operations | unlimited |
| `repos[].name` | Name used in the bundle URL | required |
| `repos[].url` | Git remote URL | required |
| `repos[].interval` | How often to fetch and re-bundle | `5m` |
| `repos[].env` | Extra environment variables for git commands | none |
| `repos[].repack` | Run `git repack -adb` before bundling | `false` |
| `repos[].clone_flags` | Extra flags passed to `git clone --bare` | none |
| `repos[].fetch_flags` | Flags passed to `git fetch` | `["--all"]` |
| `repos[].bundle_flags` | Flags passed to `git bundle create` | `["--all"]` |

Environment variables can be referenced in the config using `${env:NAME}` syntax.
