Name: hcu-container-toolkit
Version: %{version}
Release: %{release}
Group: Development Tools

Vendor: HCUOpt CORPORATION
Packager: HCUOpt CORPORATION

Summary: HCU Container Toolkit
License: Apache-2.0

Source0: hcu-ctk
Source1: hcu-container-runtime
Source2: hcu-cdi-hook
Source3: LICENSE
Source4: hcu-docker

%description
Provides tools and utilities to enable HCU support in containers.

%prep
cp %{SOURCE0} %{SOURCE1} %{SOURCE2} %{SOURCE3} %{SOURCE4} .

%install
mkdir -p %{buildroot}%{_bindir}
install -m 755 -t %{buildroot}%{_bindir} hcu-container-runtime
install -m 755 -t %{buildroot}%{_bindir} hcu-ctk
install -m 755 -t %{buildroot}%{_bindir} hcu-cdi-hook
install -m 755 -t %{buildroot}%{_bindir} hcu-docker

%posttrans
if [ ! -e %{_bindir}/nvidia-container-runtime-hook ]; then
  # repairing lost file nvidia-container-runtime-hook
  ln -sf %{_bindir}/hcu-container-runtime %{_bindir}/nvidia-container-runtime-hook
fi

# Generate the default config; If this file already exists no changes are made.
%{_bindir}/hcu-ctk --quiet config --config-file=%{_sysconfdir}/hcu-container-runtime/config.toml --in-place

%{_bindir}/hcu-ctk runtime configure --runtime=docker --set-as-default
if systemctl status docker >/dev/null 2>&1; then
  echo -e "\e[032m ================================== \e[0m"
  echo -e "\e[033m Please restart Docker service:     \e[0m"
  echo -e "\e[033m    sudo systemctl restart docker   \e[0m"
  echo -e "\e[032m ================================== \e[0m"
fi

%postun
if [ "$1" = 0 ]; then  # package is uninstalled, not upgraded
  if [ -L %{_bindir}/nvidia-container-runtime-hook ]; then rm -f %{_bindir}/nvidia-container-runtime-hook; fi
  if [ -e %{_bindir}/sdocker ]; then mv %{_bindir}/sdocker %{_bindir}/docker; fi
fi

%files
%license LICENSE
%{_bindir}/hcu-ctk
%{_bindir}/hcu-container-runtime
%{_bindir}/hcu-cdi-hook
%{_bindir}/hcu-docker
%changelog
* %{release_date} HCUOpt CORPORATION %{version}-%{release}
- See CHANGELOG.md
