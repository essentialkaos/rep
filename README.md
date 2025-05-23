<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/r/rep"><img src="https://kaos.sh/r/rep.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/y/ek"><img src="https://kaos.sh/y/ba1bd149e31f4a00abf72ac930aedac9.svg" alt="Codacy badge" /></a>
  <a href="https://kaos.sh/w/rep/ci-push"><img src="https://kaos.sh/w/rep/ci-push.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/rep/codeql"><img src="https://kaos.sh/w/rep/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#usage-demo">Usage demo</a> • <a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`rep` is a DNF/YUM repository management utility.

### Usage demo

[![demo](https://github.com/user-attachments/assets/b46c3b2a-785c-4131-a8bd-a712800b4899)](#usage-demo)

### Installation

#### From [ESSENTIAL KAOS Public Repository](https://kaos.sh/kaos-repo)

```bash
sudo dnf install -y https://pkgs.kaos.st/kaos-repo-latest.el$(grep 'CPE_NAME' /etc/os-release | tr -d '"' | cut -d':' -f5).noarch.rpm
sudo dnf install rep
```

#### Containers

Official `rep` images available on [GitHub Container Registry](https://kaos.sh/p/rep) and [Docker Hub](https://kaos.sh/d/rep). Install the latest version of [Podman](https://podman.io/getting-started/installation.html) or [Docker](https://docs.docker.com/engine/install/), then:

```bash
curl -fL# -o rep-container https://kaos.sh/rep/rep-container
chmod +x rep-container
sudo mv rep-container /usr/bin/rep

mkdir /opt/rep
export REP_DIR=/opt/rep

# Create repository configuration in /opt/rep/conf (use common/repository.knf.example as an example)

rep init src x86_64
rep add my-package.el8.x86_64.rpm
```

Official Docker images with `rep`:

- [`ghcr.io/essentialkaos/rep:latest`](https://kaos.sh/p/rep)
- [`essentialkaos/rep:latest`](https://kaos.sh/d/rep)

### Usage

<img src=".github/images/usage.svg" />

### CI Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://kaos.sh/w/rep/ci-push.svg?branch=master)](https://kaos.sh/w/rep/ci-push?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/rep/ci-push.svg?branch=develop)](https://kaos.sh/w/rep/ci-push?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/.github/blob/master/CONTRIBUTING.md).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://kaos.dev"><img src="https://raw.githubusercontent.com/essentialkaos/.github/refs/heads/master/images/ekgh.svg"/></a></p>
