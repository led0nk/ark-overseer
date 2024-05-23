%global goipatch github.com/led0nk/ark-clusterinfo

Name: ark-clusterinfo
Version: 0.1.0
Release:        1.20240523105647044231.discord%{?dist}
Summary:  steam observation tool

License:  BSD
#URL:  github.com/led0nk/ark-clusterinfo
Source0: %{name}-%{version}.tar.gz

BuildRequires: golang

%description
ark-clusterinfo is a steam observation tool to track players


%prep
mkdir -p _build/bin

%setup -q -n %{name}-%{version}

%build
go build -v -o %{gobuilddir}/bin/%{name}

%install
install -Dpm 0755 %{gobuilddir}/bin/* %{buildroot}%{_bindir}/%{name}
install -Dpm 644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service

%files
%dir %{_sysconfdir}/%{name}
%{_bindir}/%{name}
%{_unitdir}/%{name}.service

%changelog
%autochangelog

