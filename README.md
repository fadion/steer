# Steer

A deployment tool that makes it as easy and hassle-free as it can be. No matter your skill level, it only requires a minute to setup, with no complicated configurations or special permissions.

Steer relies on Git to keep track of what has changed in the project, relieving itself from the heavy burden of version control. Files are pushed on the server either via FTP or SFTP, supporting basically any host, from shared to dedicated or cloud. But it's not just a tool that reads the file tree and uploads. It supports multiple servers, deployment previews, versioning and quite some more.

## How it Works

A config file is saved in the root of the local project that holds the configuration data: server address, credentials, paths, branch, etc. Steer reads the config, connects to the server and searches for a file named `.steer-revision`. It is that remote revision file that holds the latest deployment commit. When deploying, the list of changed files are retrieved using git, so the project you're working on needs to be a git repository.

## But Why?

While there are plenty of deployment tools that try to do a lot, not everyone needs all of their features. Teams of developers will obviously have a great setup, with jobs, hooks, commands and so on. The rest, freelancers, hobbyists or solo developers just want to get the job done and not worry about infrustructure. Steer doesn't try to do more than it should and with its simple approach, it lets the developer worry about code.

### Installation

This is still a pre-release. I'm working on setting up a good testing environment, finding a good way to deliver binaries and self upgrading. All of those should hopefully happen within a couple of weeks. For now, you can only `go get github.com/fadion/steer` and `go install` it yourself.

## Setup

Before doing a deployment, you'll need a very simple `.steer` file in the root of your project that holds the configuration. The fastest way to create it is by running:

```
steer init
```

That command will create a template with some sensible defaults that you can edit with your own data. Although it doesn't look like it, in fact it's an `.ini` file. Below is an exhaustive example configuration.

```
[production]
scheme = ftp
host = ftp.example.com
port = 21
username = user
password = secret
privatekey = /path/to/key
path = /
branch = master
versions = false
logger = false
include = file.js, folder
exclude = file.css, file.html

[staging]
scheme = sftp
host = example.com
port = 22
username = staging
password = secret
```

What you should worry right now is filling up the `scheme` (ftp or sftp), `host`, `port`, `username` and `password`, so you can connect to your server. The `path` option defines the root of the deployment, which in most cases should be `/`, `public`, or something similar. The `branch` option sets the branch of the repository you want to push to. The rest of the options we'll explore later.

The names of the sections (`production` and `staging` in the above example) are important as they can be referred to while running commands. Steer supports a configuration with multiple servers and can even deploy to them all at once.

### FTP

FTP needs the `host`, `port` (usually: 21), `username`, `password` and an absolute `path` to the root folder of your project.

### SFTP

For SFTP you can user either a combination of `username` and `password`, or a password-less authentication by setting a `privatekey` with the path to the private key. The `path` may be relative to the remote user's base directory or absolute.

## Deploy

With the configuration ready, nothing stops you from going hot. Just run:

```
steer deploy
```

The deployment process will read your git repository for a file list, prepare them and start uploading to the server. If you have a lot of files, especially in the first run, it will take a while so sit back and relax while it does its job.

By default, steer will deploy to the first server it finds in the configuration. You can change this behaviour by passing one or more server names as arguments:

```
steer deploy production staging
```

It can even deploy to all the servers at once:

```
steer deploy --all
```

There may be rare cases when you won't need to deploy the working tree, but a specific commit in the past. Maybe you're still working on a feature and were too lazy to create a branch or the update introduced some regression. If that's the case, you can pass a commit hash as an option:

```
steer deploy --commit=SOMEHASH
steer deploy -c=SOMEHASH
```

Finally, if you want to discard everything that's in the server and start fresh, there's an option for that. Please be aware that it will override every file on the server, so think if it's what you want.

```
steer deploy --fresh
```

The `commit` and `fresh` options can be used alongside server arguments:

```
steer deploy production -c=SOMEHASH
```

## Preview

Before running a deploy, it's generally not a bad idea to do a preview run. This command gets the revision, calculates which files have changed and only displays them, without uploading anything on the server. It's especially useful on big updates.

```
steer preview
```

As with deploys, you can pass a list of servers as arguments, and the `--all` and `--commit` options.

```
steer preview staging -c=SOMEHASH
```

## Sync

Deployment will read the remote revision file and push the files that were changed since that commit. Sometimes however, you may need to manually update the remote revision. Maybe you've made some quick changes, both locally and on the server, so there's no need to deploy those files. More probably, you may hear of steer while some of your projects are online. In that case, you can sync the project's current state, without the need to do deploy all the files.

```
steer sync
```

With the command executed, the remote revision will hold the latest commit and be in sync with the local repository. You can also pass a list of servers as arguments, the `--all` and `--commit` options.

```
steer sync --all
```

## Versions

Steer can behave in a way vaguely similar to deployment tools like Capistrano or Rocketeer, albeit much more lightweight. When versions is activated, steer will no longer deploy to the base path as usual, but instead inside a `versions` directory. Every deploy will create a new directory that holds the whole project. Example:

```
/versions
   /version-1111
   /version-2222
   /version-3333
```

Obviously the version folders will have a timestamp appended, ensuring their uniqueness. Once you've uploaded a new version, the general practice is to symlink that version with the project. That way, you ensure zero downtime and can test new versions in a production environment.

To activate versions, you have to enable a `versions` configuration option. The default directory is `versions` and that should be a good name for most people. However, if for some reason you want to change it, there's also the `vfolder` option. It must be set relative to the `path` option. Make sure to manually create the `versions` directory on the server, otherwise it will fail.

```
[production]
; ...
versions = true
vfolder = versiones
```

Keep in mind that versions will upload the whole project, disregarding revisions, so it will take a while depending on the amount of files. Think before going on the versions route if that's what you want and if the benefits outweigh the simplicity of regular deployments.

Versions can be enabled or disabled interactively on the command line. These help in switching the type of deploy without touching the configuration.

Enable versions:

```
steer deploy --verions
```

Disable versions:

```
steer deploy --no-verions
```

## Logging

A simple logger is available that writes on the server a `.steer-log` with information about the deployment: date and time, commit and number of changed files. By default it's disabled, but can be easily enabled by setting a `logger` configuration option:

```
[production]
; ...
versions = true
vfolder = versiones
```

Deployments can have custom messages attached to the log:

```
steer deploy -m="Updated the front-end"
```

Additionally, steer offers a few commands to work with the log file. You can get the latest line from the log by running:

```
steer log
```

If you need a few more lines, you can pass the `latest` option:

```
steer log --latest=3
```

Finally, you can even delete the log file completely by typing:

```
steer log --clear
```

## File Includes and Excludes

File includes are files or directories that you want to include in the deployment, even though they aren't tracked by git. File excludes are the opposite: files or directories that may be tracked by git, but you don't want to deploy. Directories aren't read recursively, meaning that only the files inside a specified directory will be added or removed, not the contents of sub directories.

To include or exclude files, there are appropriately named configuration options. Files and directories should be separated by a comma and be relative to the project's root directory.

```
[production]
; ...
include = file1.css, file2.js, vendor
exclude = css/vendor.css
```

## Getting Help

Steer's commands and options are well documented and most of the time, you won't need any more documentation. For general help type:

```
steer help
```

For information on a specific command and it's options, type:

```
steer deploy -h
```

## Credits

Steer was developed by Fadion Dashi, a freelance web and mobile developer from Tirana.

Most of the inspiration for this project comes from [PHPloy](https://github.com/banago/PHPloy), a deployment tool created by a dear friend, in which I've also been involved. I'm aiming feature parity and even added some of my own ideas.