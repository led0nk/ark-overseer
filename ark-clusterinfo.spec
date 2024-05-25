%global goipath github.com/led0nk/ark-clusterinfo

%define debug_package %{nil}

Version: 0.1.0

%gometa

Name: ark-clusterinfo
Release:  1%{?dist}
Summary:  steam observation tool

License:  BSD
Source0: %{name}-%{version}.tar.gz

BuildRequires: golang
BuildRequires: make
BuildRequires: git

%description
ark-clusterinfo is a steam observation tool to track players


%prep
%autosetup

%build
go build -v -buildmode pie -mod vendor -o %{gobuilddir}/bin/%{name}/cmd/server/main.go

%install
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}

install -Dpm 0755 %{gobuilddir}/bin/* %{buildroot}%{_bindir}/%{name}
install -Dpm 0644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service

%check
%gocheck

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service

%changelog
%autochangelog

