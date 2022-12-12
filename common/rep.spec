################################################################################

# rpmbuilder:relative-pack true

################################################################################

%global crc_check pushd ../SOURCES ; sha512sum -c %{SOURCE100} ; popd

################################################################################

%define debug_package %{nil}

################################################################################

%define _opt     /opt
%define _logdir  %{_localstatedir}/log

################################################################################

Summary:         YUM repository management utility
Name:            rep
Version:         3.0.3
Release:         0%{?dist}
Group:           Applications/System
License:         Apache 2.0
URL:             https://kaos.sh/rep

Source0:         https://source.kaos.st/%{name}/%{name}-%{version}.tar.bz2

Source100:       checksum.sha512

BuildRoot:       %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:   golang >= 1.19

Requires:        createrepo_c

Provides:        %{name} = %{version}-%{release}

################################################################################

%description
YUM repository management utility.

################################################################################

%prep
%{crc_check}

%setup -q

%build
export GOPATH=$(pwd)

if [[ ! -d "%{name}/vendor" ]] ; then
  echo "This package requires vendored dependencies"
  exit 1
fi

pushd %{name}
  %{__make} %{?_smp_mflags} all
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/%{name}.d
install -dm 750 %{buildroot}%{_localstatedir}/cache/%{name}
install -dm 755 %{buildroot}%{_logdir}/%{name}
install -dm 755 %{buildroot}%{_mandir}/man1

install -dm 755 %{buildroot}%{_opt}/%{name}

install -pm 755 %{name}/%{name} \
                %{buildroot}%{_bindir}/

install -pm 644 %{name}/common/%{name}.knf \
                %{buildroot}%{_sysconfdir}/
install -pm 644 %{name}/common/*.example \
                %{buildroot}%{_sysconfdir}/%{name}.d/

./%{name}/%{name} --generate-man > %{buildroot}%{_mandir}/man1/%{name}.1

%clean
rm -rf %{buildroot}

%post
if [[ -d %{_sysconfdir}/bash_completion.d ]] ; then
  %{name} --completion=bash 1> %{_sysconfdir}/bash_completion.d/%{name} 2>/dev/null
fi

if [[ -d %{_datarootdir}/fish/vendor_completions.d ]] ; then
  %{name} --completion=fish 1> %{_datarootdir}/fish/vendor_completions.d/%{name}.fish 2>/dev/null
fi

if [[ -d %{_datadir}/zsh/site-functions ]] ; then
  %{name} --completion=zsh 1> %{_datadir}/zsh/site-functions/_%{name} 2>/dev/null
fi

%postun
if [[ $1 == 0 ]] ; then
  if [[ -f %{_sysconfdir}/bash_completion.d/%{name} ]] ; then
    rm -f %{_sysconfdir}/bash_completion.d/%{name} &>/dev/null || :
  fi

  if [[ -f %{_datarootdir}/fish/vendor_completions.d/%{name}.fish ]] ; then
    rm -f %{_datarootdir}/fish/vendor_completions.d/%{name}.fish &>/dev/null || :
  fi

  if [[ -f %{_datadir}/zsh/site-functions/_%{name} ]] ; then
    rm -f %{_datadir}/zsh/site-functions/_%{name} &>/dev/null || :
  fi
fi

################################################################################

%files
%defattr(-,root,root,-)
%doc %{name}/LICENSE
%config(noreplace) %{_sysconfdir}/%{name}.knf
%dir %{_localstatedir}/cache/%{name}
%dir %{_opt}/%{name}
%dir %{_logdir}/%{name}
%{_mandir}/man1/%{name}.1.*
%{_sysconfdir}/%{name}.d/*.example
%{_bindir}/%{name}

################################################################################

%changelog
* Mon Dec 12 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.4-0
- Fixed bug with re-signing all packages
- Added packages prefiltering to 'add' command

* Tue Oct 11 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.2-0
- Added logging for re-signing packages

* Mon Oct 10 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.1-0
- Fixed bug with filtering packages by release status

* Mon Jun 27 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.0-0
- First public release of 3.x
