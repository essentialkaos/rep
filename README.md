<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/r/rep"><img src="https://kaos.sh/r/rep.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/l/rep"><img src="https://kaos.sh/l/5876fdc611100e9f8a83.svg" alt="Code Climate Maintainability" /></a>
  <a href="https://kaos.sh/b/rep"><img src="https://kaos.sh/b/07867ea4-6025-47a8-ad18-112dd7b37a3c.svg" alt="codebeat badge" /></a>
  <a href="https://kaos.sh/w/rep/ci"><img src="https://kaos.sh/w/rep/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/rep/codeql"><img src="https://kaos.sh/w/rep/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#usage-demo">Usage demo</a> • <a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#license">License</a></p>

<br/>

`rep` is a DNF/YUM repository management utility.

### Usage demo

[![demo](https://gh.kaos.st/rep-300.gif)](#usage-demo)

### Installation

#### From [ESSENTIAL KAOS Public Repository](https://kaos.sh/kaos-repo)

```bash
sudo yum install -y https://pkgs.kaos.st/kaos-repo-latest.el$(grep 'CPE_NAME' /etc/os-release | tr -d '"' | cut -d':' -f5).noarch.rpm
sudo yum install rep
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
rep add my-package.el7.x86_64.rpm
```

Official Docker images with `rep`:

- [`ghcr.io/essentialkaos/rep:latest`](https://kaos.sh/p/rep)
- [`essentialkaos/rep:latest`](https://kaos.sh/d/rep)

### Usage

<img src=".github/images/usage.svg" />

### CI Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://kaos.sh/w/rep/ci.svg?branch=master)](https://kaos.sh/w/rep/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/rep/ci.svg?branch=develop)](https://kaos.sh/w/rep/ci?query=branch:develop) |

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
