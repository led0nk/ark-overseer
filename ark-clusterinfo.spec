%global goipath github.com/led0nk/ark-clusterinfo
%global forgeurl https://github.com/led0nk/ark-clusterinfo

%define debug_package %{nil}

Version: 0.1.0

%gometa

Name: %{goname}
Release:  1%{?dist}
Summary:  steam observation tool

License:  BSD
Source0: %{gosource}

BuildRequires: systemd-rpm-macros
BuildRequires: go-rpm-macros
BuildRequires: golang
BuildRequires: git
BuildRequires: make

%description
ark-clusterinfo is a steam observation tool to track players

%gopkg

%prep
%goprep
%autosetup 

%generate_buildrequires
%go_generate_buildrequires

%build
for cmd in cmd/* ; do
  %gobuild -o %{gobuilddir}/bin/$(basename $cmd) %{goipath}/$cmd
done
#go build -v -buildmode pie -mod vendor -o %{gobuilddir}/bin/%{name} cmd/server/main.go

%install
%gopkginstall
install -m 0755 -vd                     %{buildroot}%{_bindir}
#install -m 0755 -vd                     %{buildroot}%{_unitdir}
install -m 0755 -vp %{gobuilddir}/bin/* %{buildroot}%{_bindir}/
#install -m 0644 -vp %{name}.service %{buildroot}%{_unitdir}/

%check
%gocheck

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%files
%{_bindir}/%{name}
#%{_unitdir}/%{name}.service
%gopkgfiles

%changelog
%autochangelog

