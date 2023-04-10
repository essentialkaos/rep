<p align="center"><a href="#readme"><img src="https://gh.kaos.st/rep.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/rep/ci"><img src="https://kaos.sh/w/rep/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/r/rep"><img src="https://kaos.sh/r/rep.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/b/rep"><img src="https://kaos.sh/b/07867ea4-6025-47a8-ad18-112dd7b37a3c.svg" alt="codebeat badge" /></a>
  <a href="https://kaos.sh/w/rep/codeql"><img src="https://kaos.sh/w/rep/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#usage-demo">Usage demo</a> • <a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#license">License</a></p>

<br/>

`rep` is a YUM repository management utility.

### Usage demo

[![demo](https://gh.kaos.st/rep-300.gif)](#usage-demo)

### Installation

#### From [ESSENTIAL KAOS Public Repository](https://yum.kaos.st)

```bash
sudo yum install -y https://yum.kaos.st/kaos-repo-latest.el$(grep 'CPE_NAME' /etc/os-release | tr -d '"' | cut -d':' -f5).noarch.rpm
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

```
Usage: rep {options} {command}

Notice that if you have more than one repository you should define its name as
the first argument. You can read detailed info about every command with usage
examples using help command.

Commands

  init arch…              Initialize new repository
  gen-key                 Generate keys for signing packages
  list filter             List latest versions of packages within repository
  find query…             Search packages
  which-source query…     Show source package name
  info package            Show info about package
  payload package type    Show package payload
  cleanup num             Remove old versions of packages
  check errors-num        Check repositories consistency
  sign file…              Sign one or more packages
  resign                  Resign all packages in repository
  add file…               Add one or more packages to testing repository
  remove query…           Remove package or packages from repository
  release query…          Copy package or packages from testing to release repository
  unrelease query…        Remove package or packages from release repository
  reindex                 Create or update repository index
  purge-cache             Clean all cached data
  stats                   Show some statistics information about repositories
  help command            Show detailed information about command

Options

  --release, -r           Run command only on release (stable) repository
  --testing, -t           Run command only on testing (unstable) repository
  --all, -a               Run command on all repositories
  --arch, -aa arch        Package architecture (helpful with "info" and "payload" commands)
  --move, -m              Move (remove after successful action) packages (helpful with "add" command)
  --no-source, -ns        Ignore source packages (helpful with "add" command)
  --ignore-filter, -if    Ignore repository file filter (helpful with "add" and "sign" commands)
  --force, -f             Answer "yes" for all questions
  --full, -F              Full reindex (helpful with "reindex" command)
  --show-all, -A          Show all versions of packages (helpful with "list" command)
  --status, -S            Show package status (released or not)
  --epoch, -E             Show epoch info (helpful with "list" and "which-source" commands)
  --pager, -P             Run command in "pager" mode (i.e. don't disable colors and don't show raw output)
  --no-color, -nc         Disable colors in output
  --help, -h              Show this help message
  --version, -v           Show version
```

### CI Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://kaos.sh/w/rep/ci.svg?branch=master)](https://kaos.sh/w/rep/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/rep/ci.svg?branch=develop)](https://kaos.sh/w/rep/ci?query=branch:develop) |

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
