# ocm-workspace
A containerised workspace for running OpenShift Dedicated (OSD) cluster management tools (e.g. ocm-cli, oc, etc).

# Goal
Provide a deployable, extensible and isolated (container-isolation only) workspace for managing OSD clusters.

> **Important:** ocm-workspace is a user container workspace which is intended to be customized by the user only and not a means to control "how to do" things.

# Prerequisites
1. Clone this repository where the branch must be newer than the tag `1.0.0-beta`.
2. Installed podman, golang and make binaries.
3. A configuration file located in `/home/<user>/.ocm-workspace.yaml`. See [configuration](#configuration)


# Build, Install, and Build the Workspace Image
To install the ocm-workspace binary in your $GOPATH, run the following.

```
$ cd ocm-workspace
$ make all
```

# Log into an OSD cluster
To run by logging into an OSD cluster run the following.

```
$ workspace login -c <cluster_name or id>
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


# Mimimum Required Configuration
**Prerequisites**
1. `.ocm-workspace.yaml` in `/home/<user>/.ocm-workspace.yaml`
2. `config.<prod|stage>.json` in `/home/<user>/.config/backplane`

For more information see [config.md](./config.md).

**Steps**
1. Create the expected configuration file.
2. Configure the minimum items like the following.

```
# OCM/OSD
ocUser: <OCM user name>
ocmEnvironment: <production or staging>
backplaneConfigProd: "<production backplane config file path>"

# Host
userHome: "<Path to user's home directory.>"

# container build params
baseImage: "fedora:37"
ocmCLIVersion: "0.1.60"
rhocCLIVersion: "0.0.37"
backplaneCLIVersion: "0.1.2"
```

# Plugins
The workspace supports running of plugins at certain points in its execution. Plugins are executables that run inside the ocm-workspace container but reside in a different repository including their binaries that need to be built by a user before it can be used by ocm-workspace. A plugin's executable is mounted at container run as a volume.

### portForward Plugin
An example of a working plugin called `portForward` that automatically forwards ports of a kubernetes service, to a container port and finally to a host port so that a user can access the k8s service from the host. For example, this plugin can be used to access some web console (e.g. prometheus console) from the host.

A user must configure their plugins and ensure they can be located by ocm-workspace.
```
plugins:
  - name: portForward
    execPath: /home/jcueto/workspace/repos/ocm-workspace-plugins/portForward/portForward
    runOn: ocmBackplaneLoginSuccess
    allocatePorts: 2
    execCommand: portForward
    config: |
      services:
        - name: rhoam
          ports:
            - name: prometheus-console
              namespace: redhat-rhoam-operator-observability
              source:
                kind: pod
                name: prometheus-rhoam-0
              svcPort: 9090
            - name: alertmanager-console
              namespace: redhat-rhoam-operator-observability
              source:
                kind: pod
                name: alertmanager-rhoam-0
              svcPort: 9093

```

**Breakdown of Configuration**

`plugins` -  A list of plugin configurations.

`plugins.name` - The name of the plugin.

`plugins.execPath` - The path to the plugin's executable.

`plugins.runOn` - The workspace's execution point where a plugin must be started.

`plugins.allocatePorts`: This tells the workspace to allocate the number of ports that are mapped from the host to the container (e.g. hostport:containerport)

`execCommand` - This is the plugin's executable CLI command. Therefore a plugin is required to at least have one CLI command.

`plugins.config` - This is the plugin's config file that must be a YAML file. The workspace does not use this config, however it makes it available to the plugin inside the container.

