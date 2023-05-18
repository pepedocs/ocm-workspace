# ocm-workspace
A containerised workspace or environment for running OpenShift Dedicated (OSD) cluster management tools (e.g. ocm-cli, oc, etc).

# Goal
Provide a deployable and isolated (container-isolation only) environment for managing OSD clusters.

# Prerequisites
1. Clone this repository.
2. Installed podman, golang and make binaries.
3. An exactly named configuration file located in `/home/<user>/.ocm-workspace.yaml`. See [configuration](#configuration)
4. An exactly named backplane configuration file located in `/home/<user>/.config/backplane/config.<prod|stage>.json`.


# Install
To install the ocm-workspace binary, run the following.

```
$ cd ocm-workspace
$ make install
```

# Build ocm-workspace image
To build the ocm-workspace image, run the following.

```
$ cd ocm-workspace
$ make buildImage
```

# Log into an OSD cluster
To run by logging into an OSD cluster run the following.

```
$ ./workspace login -c <cluster_name or id>
```

A container will be created (see `podman ps`) and a bash terminal will be provided for running cluster management commands. The following operations are executed automatically:

- OCM Login
- OpenShift cluster login

```
[<user>@<cluster name or id> <current kubernetes namespace>]$
```

# Run ocm-workspace without logging into an OSD cluster
```
$ ./workspace login --isOcmLoginOnly
```

A container will be created and bash terminal will be provided for running cluster management commands. The following operations are executed automatically:
- OCM Login

# Launch an OpenShift console for a running ocm-workspace container
**Steps**
1. Get the running ocm-workspace container name by searching it in `podman ps`. The name is in the form of `ow-<cluster name>-uid`
2. In the running ocm-workspace container get the value of the environment variable `OPENSHIFT_CONSOLE_PORT`.
3. Run the follwing command.

```
$ ./workspace openshiftConsole -c ow-<cluster name>-uid <value of OPENSHIFT_CONSOLE_PORT>
```
4. The OpenShift console will be available in your browser.

```
http://localhost:<value of OPENSHIFT_CONSOLE_PORT>
```


# Configuration
**Prerequisites**
1. `.ocm-workspace.yaml` in `/home/<user>/.ocm-workspace.yaml`
2. `config.<prod|stage>.json` in `/home/<user>/.config/backplane`


**Steps**
1. Create the expected configuration file.
2. Configure the minimum items like the following.

```
# OCM/OSD
ocUser: <OCM user name>
ocmEnvironment: <production or staging>
backplaneConfigProd: "<BACKPLANE CONFIG FILENAME>"

# container build params
baseImage: "fedora:37"
ocmCLIVersion: "0.1.60"
rhocCLIVersion: "0.0.37"
backplaneCLIVersion: "0.1.2"
```

# Service Reference Configuration
A service reference can be added in the configuration file to enable tool features relevant to a certain service. This feature is useful for accessing a service from the ocm-workspace's host (e.g. web console). Note that from the host's side, its port is only accessible through the loopback interface.

For example the following will automatically forward the specified ports of service `nginx` to an ocm-workspace port that is also mapped to a host port.

```
# forward to host mapped ports
services:
  - name: nginx
    forwardPorts:
      - name: prometheus-console
        namespace: nginx
        source:
          kind: pod
          name: nginx-prometheus
        port: 9090
      - name: alertmanager-console
        namespace: nginx
        source:
          kind: pod
          name: nginx-alertmanager
        port: 9093
```


# Todo
1. Create tests
2. Enhance error-handling
3. Refactor
4. Hooks/Plugins

