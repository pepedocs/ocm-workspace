# ocm-workspace
A containerised workspace or environment for running OpenShift Dedicated (OSD) cluster management tools (e.g. ocm-cli, oc, etc).

# Goal
Provide a deployable and isolated (not 100%) environment for managing OSD clusters.

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

# Build ocm-workspace Image
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

A container will be created (see `podman ps`) and a bash terminal will be provided for running cluster management commands. The following is also done:

- OCM Login
- OpenShift cluster login

```
[<user>@<cluster name or id> <current kubernetes namespace>]$
```

# Run ocm-workspace Without Logging into an OSD cluster
```
$ ./workspace login --isOcmLoginOnly
```

A container will be created and bash terminal will be provided for running cluster management commands.
- OCM Login

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
ocmToken: "<OCM token>"

# container build params
baseImage: "fedora:37"
ocmCLIVersion: "0.1.60"
rhocCLIVersion: "0.0.37"
backplaneCLIVersion: "0.1.2"
```

# Todo
1. Create tests
2. Enhance error-handling
3. Refactor
4. Hooks/Plugins

