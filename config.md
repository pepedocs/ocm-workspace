# Configuration
This section describes how to configure ocm-workspace using a configuration file.

# Prerequisites
1. If the `--config` parameter is not specified, the config file is expected at the user's home directory with the name of `.ocm-workspace.yaml`.
2. The config file must be in the `YAML` format.

# Sections
There are three major sections in the config file which are described below.

# OCM
This section is for configuring `OCM` information that the workspace will/can use. The following are the available configuration properties so far:

`ocUser` - The OpenShift CLI user to use.

`ocmEnvironment` - The OCM environment to run ocm commands for.

`backplaneConfigProd` - The filename of the backplane config file for the OCM production environment.

`backplaneConfigStage` - The filename of the backplane config file for the OCM stage environment.

> Note: The backplaneConfig* files will be looked up in the user's home directory at `/path to user home/.config/backplane.`

# Workspace Settings
The workspace settings are for configuring the workspace build and run behavior.

`userHome` - The user's home directory in the host machine.

`baseImage` - The base image of the workspace image.

`ocmCLIVersion` - The version of the OCM CLI to use inside the container.

`backplaneCLIVersion` - The version of the backplane CLI to use inside the container.

`allocateFreePorts` - Setting this to N will allocate N free TCP ports that are mapped from the host to the container.

> Note: These ports can only be accessed locally (localhost) when accessed from the host.

`hostUser` - The user's username in the host machine.

`customDirMaps` - A list of `hostdir:containerdir` directory volume maps.

`addToPATHEnv` - A list of container directories that is added to the container's PATH environment variable.

`exportEnvVars` - A list of container environment variables that is exported inside the container.
