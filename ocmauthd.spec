# 
# ocmauthd spec file
#

Name: ocmauthd
Summary: Authentication daemon for CERNBox OCM implementation.
Version: 0.0.1
Release: 1%{?dist}
License: AGPLv3
BuildRoot: %{_tmppath}/%{name}-buildroot
Group: CERN-IT/ST
BuildArch: x86_64
Source: %{name}-%{version}.tar.gz

%description
This RPM provides a golang webserver that provides an authentication service for web clients.

# Don't do any post-install weirdness, especially compiling .py files
%define __os_install_post %{nil}

%prep
%setup -n %{name}-%{version}

%install
# server versioning

# installation
rm -rf %buildroot/
mkdir -p %buildroot/usr/local/bin
mkdir -p %buildroot/etc/ocmauthd
mkdir -p %buildroot/etc/logrotate.d
mkdir -p %buildroot/usr/lib/systemd/system
mkdir -p %buildroot/var/log/ocmauthd
install -m 755 ocmauthd	     %buildroot/usr/local/bin/ocmauthd
install -m 644 ocmauthd.service    %buildroot/usr/lib/systemd/system/ocmauthd.service
install -m 644 ocmauthd.yaml       %buildroot/etc/ocmauthd/ocmauthd.yaml
install -m 644 ocmauthd.logrotate  %buildroot/etc/logrotate.d/ocmauthd

%clean
rm -rf %buildroot/

%preun

%post

%files
%defattr(-,root,root,-)
/etc/ocmauthd
/etc/logrotate.d/ocmauthd
/var/log/ocmauthd
/usr/lib/systemd/system/ocmauthd.service
/usr/local/bin/*
%config(noreplace) /etc/ocmauthd/ocmauthd.yaml


%changelog
* Wed Oct 10 2018 Diogo Castro <diogo.castro@cern.ch> 0.0.1
- v0.0.1

