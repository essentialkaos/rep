################################################################################

# rpmbuilder:relative-pack true

################################################################################

%global crc_check pushd ../SOURCES ; sha512sum -c %{SOURCE100} ; popd

################################################################################

%define _posixroot        /
%define _root             /root
%define _bin              /bin
%define _sbin             /sbin
%define _srv              /srv
%define _home             /home
%define _opt              /opt
%define _lib32            %{_posixroot}lib
%define _lib64            %{_posixroot}lib64
%define _libdir32         %{_prefix}%{_lib32}
%define _libdir64         %{_prefix}%{_lib64}
%define _logdir           %{_localstatedir}/log
%define _rundir           %{_localstatedir}/run
%define _lockdir          %{_localstatedir}/lock/subsys
%define _cachedir         %{_localstatedir}/cache
%define _spooldir         %{_localstatedir}/spool
%define _crondir          %{_sysconfdir}/cron.d
%define _loc_prefix       %{_prefix}/local
%define _loc_exec_prefix  %{_loc_prefix}
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_libdir       %{_loc_exec_prefix}/%{_lib}
%define _loc_libdir32     %{_loc_exec_prefix}/%{_lib32}
%define _loc_libdir64     %{_loc_exec_prefix}/%{_lib64}
%define _loc_libexecdir   %{_loc_exec_prefix}/libexec
%define _loc_sbindir      %{_loc_exec_prefix}/sbin
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_datarootdir  %{_loc_prefix}/share
%define _loc_includedir   %{_loc_prefix}/include
%define _loc_mandir       %{_loc_datarootdir}/man
%define _rpmstatedir      %{_sharedstatedir}/rpm-state
%define _pkgconfigdir     %{_libdir}/pkgconfig

################################################################################

%define debug_package %{nil}

################################################################################

%define srcdir src/github.com/essentialkaos/%{name}

################################################################################

Summary:         YUM repository management utility
Name:            rep
Version:         3.0.0
Release:         0%{?dist}
Group:           Applications/System
License:         Apache 2.0
URL:             https://kaos.sh/rep

Source0:         https://source.kaos.st/%{name}/%{name}-%{version}.tar.bz2

Source100:       checksum.sha512

BuildRoot:       %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:   golang >= 1.17

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

pushd %{srcdir}
  %if 0%{?unstable}
    %{__make} %{?_smp_mflags} all GITREV=$(cat .REVISION)
  %else
    %{__make} %{?_smp_mflags} all
  %endif
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/%{name}.d
install -dm 750 %{buildroot}%{_localstatedir}/cache/%{name}
install -dm 755 %{buildroot}%{_logdir}/%{name}

install -dm 755 %{buildroot}%{_opt}/%{name}

install -pm 755 %{srcdir}/%{name} \
                %{buildroot}%{_bindir}/

install -pm 644 %{srcdir}/common/%{name}.knf \
                %{buildroot}%{_sysconfdir}/
install -pm 644 %{srcdir}/common/*.example \
                %{buildroot}%{_sysconfdir}/%{name}.d/

%clean
rm -rf %{buildroot}

################################################################################

%files
%defattr(-,root,root,-)
%doc LICENSE
%config(noreplace) %{_sysconfdir}/%{name}.knf
%dir %{_localstatedir}/cache/%{name}
%dir %{_opt}/%{name}
%dir %{_logdir}/%{name}
%{_sysconfdir}/%{name}.d/*.example
%{_bindir}/%{name}

################################################################################

%changelog
* Mon Apr 11 2022 Anton Novojilov <andy@essentialkaos.com> - 3.0.0-0
- First public release of 3.x
