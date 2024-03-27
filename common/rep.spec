################################################################################

%global crc_check pushd ../SOURCES ; sha512sum -c %{SOURCE100} ; popd

################################################################################

%define debug_package %{nil}

################################################################################

%define _opt     /opt
%define _logdir  %{_localstatedir}/log

################################################################################

Summary:        YUM repository management utility
Name:           rep
Version:        3.3.4
Release:        0%{?dist}
Group:          Applications/System
License:        Apache 2.0
URL:            https://kaos.sh/rep

Source0:        https://source.kaos.st/%{name}/%{name}-%{version}.tar.bz2

Source100:      checksum.sha512

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:  golang >= 1.22

Requires:       createrepo_c

Provides:       %{name} = %{version}-%{release}

################################################################################

%description
YUM repository management utility.

################################################################################

%prep
%{crc_check}

%setup -q

%build
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

install -dm 755 %{buildroot}%{_sysconfdir}/bash_completion.d
install -dm 755 %{buildroot}%{_datadir}/zsh/site-functions
install -dm 755 %{buildroot}%{_datarootdir}/fish/vendor_completions.d

install -pm 755 %{name}/%{name} \
                %{buildroot}%{_bindir}/

install -pm 644 %{name}/common/%{name}.knf \
                %{buildroot}%{_sysconfdir}/
install -pm 644 %{name}/common/*.example \
                %{buildroot}%{_sysconfdir}/%{name}.d/

./%{name}/%{name} --generate-man > %{buildroot}%{_mandir}/man1/%{name}.1

./%{name}/%{name} --completion=bash 1> %{buildroot}%{_sysconfdir}/bash_completion.d/%{name}
./%{name}/%{name} --completion=zsh 1> %{buildroot}%{_datadir}/zsh/site-functions/_%{name}
./%{name}/%{name} --completion=fish 1> %{buildroot}%{_datarootdir}/fish/vendor_completions.d/%{name}.fish

%clean
rm -rf %{buildroot}

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
%{_sysconfdir}/bash_completion.d/%{name}
%{_datadir}/zsh/site-functions/_%{name}
%{_datarootdir}/fish/vendor_completions.d/%{name}.fish

################################################################################

%changelog
* Thu Mar 21 2024 Anton Novojilov <andy@essentialkaos.com> - 3.3.4-0
- Improved support information gathering
- Code refactoring
- Dependencies update

* Wed Mar 06 2024 Anton Novojilov <andy@essentialkaos.com> - 3.3.3-0
- Fixed minor bug with rendering packages list
- Code refactoring
- Dependencies update

* Mon Jan 29 2024 Anton Novojilov <andy@essentialkaos.com> - 3.3.2-0
- Dependencies update

* Tue Dec 19 2023 Anton Novojilov <andy@essentialkaos.com> - 3.3.1-0
- Dependencies update

* Mon Oct 09 2023 Anton Novojilov <andy@essentialkaos.com> - 3.3.0-0
- Added -pi/--postpone-index option to postpone index rebuild after some
  commands
- Added package filtering for 'cleanup' command
- Improved pagers (more/less) support
- UI improvements
- Fixed 'payload' command output

* Thu Oct 05 2023 Anton Novojilov <andy@essentialkaos.com> - 3.2.0-0
- Added xz compression support for repository metadata
- Added zst compression support for repository metadata
- Added changelog date to 'info' command output
- Improved changelog record search
- Fixed bug with using compression type defined in configuration file

* Tue Jun 27 2023 Anton Novojilov <andy@essentialkaos.com> - 3.1.2-0
- Minor UI fix
- Dependencies update

* Mon Jun 26 2023 Anton Novojilov <andy@essentialkaos.com> - 3.1.1-0
- Added pauses between checks in 'check' command output
- Fixed panic while checking repositories consistency
- Fixed some typos
- Dependencies update

* Sun Mar 12 2023 Anton Novojilov <andy@essentialkaos.com> - 3.1.0-0
- Added 'cleanup' command

* Mon Dec 12 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.3-0
- Fixed bug with re-signing all packages
- Added packages prefiltering to 'add' command

* Tue Oct 11 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.2-0
- Added logging for re-signing packages

* Mon Oct 10 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.1-0
- Fixed bug with filtering packages by release status

* Mon Jun 27 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.0-0
- First public release of 3.x
