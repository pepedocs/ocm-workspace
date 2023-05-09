# ocm-workspace
A containerised workspace or environment for running OpenShift Dedicated (OSD) cluster management tools (e.g. ocm-cli, oc, etc).

# Goal
Provide a deployable and isolated (not 100%) environment for managing OSD clusters.

# Prerequisites
1. Clone this repository.
2. Installed podman, golang and make binaries.
3. An exactly named configuration file located in `/home/<user>/.ocm-workspace.yaml`. See [configuration](#configuration) for more information on creating the barebones configuration.

# Install
To install run the following.

`$ cd ocm-workspace && make install`

The binary called `workspace` will be built on the current directory.

# Run
To run by logging into an OSD cluster run the following.

`$ ./workspace login -c <cluster_name or id>`

A container will be created (see `podman ps`) and a bash terminal will be provided for running cluster management commands.

`[<user>@<cluster name or id> <current kubernetes namespace>]$ `


# Configuration
ocm-workspace expects a file named `.ocm-workspace.yaml` in a user's home directory - `/home/<user>/.ocm-workspace.yaml`.

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
1. Tests
2. Enhance error-handling
3. Refactor
