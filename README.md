# Steer

Steer is a deployment tool that relies on Git to keep track of what has changed in the project. Using the raw performance and concurrency of Go, it pushes files in parallel on either a FTP or SFTP server. But it's not just a tool that reads the file tree and uploads. It supports multiple servers, previews, atomic deploys and quite some more.

## Table of Contents

- [Installation](#installation)
- [Setup](#setup)
- [Deploy](#deploy)
- [Parallel Operations](#parallel-operations)
- [Preview](#preview)
- [Sync](#sync)
- [Atomic Deployments](#atomic-deployments)
- [Logging](#logging)
- [Pre and Post Deployment Commands](#pre-and-post-deployment-commands)
- [File Includes and Excludes](#file-includes-and-excludes)
- [Getting Help](#getting-help)
- [Credits](#credits)

## Installation

Steer is still a pre-release software. I'm trying to manually test it in as many scenarios as possible and setting up an automated testing environment.

### Download from the Source

Open up the [latest release](https://github.com/fadion/steer/releases/latest) and download one of the archives corresponding to your operating system and architecture. If you're on a 64-bit macOS for example, you'll need "steer-macos-64bit.zip". Extract the archive, move it in a location under $PATH (or %PATH% on Windows) and you're good to go.
 
### Updates

Steer has self-updating capabilities, so you'll need to install it only once and then update just by running:

```
steer update
```

If there's a new version, Steer will download it and replace your existing binary.

### Using the Code

If you want to make a pull request or just edit the code for fun, you can build Steer as usual. It needs at least `go 1.8` with GOPATH configured and $GOPATH/bin in your path to run the executable.

```
go get github.com/fadion/steer
go install
steer version
```

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
atomic = false
logger = false
include = file.js, folder
exclude = file.css, file.html
maxclients = 3
predeploy = rm -rf cache
postdeploy = npm update, gulp

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

## Parallel Operations

Doing a single operation synchronously would make deployment quite a slow process, especially when a lot of files are involved. Fortunately, Steer can upload and delete files in parallel on both FTP and SFTP, speeding up the process substantially. The number of concurrent operations varies from the server configuration, so you may start with a sensible number like 3 (the default) and increase it until you notice errors while deploying.

The `maxclients` configuration option sets the maximum number of concurrent operations. Set to 1 for no parallelism.

```
[production]
; ...
maxclients = 5
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

## Atomic Deployments

When high availability of your site is critical, you can use atomic deployments for virtually no down time during updates. Steer will no longer update files in the base path, but instead push the whole project inside a `releases` directory. To get an idea:

```
/current
/releases
   /111111111
   /222222222
   /333333333
```

Each deployment will create a new directory to ensure uniqueness, holding every file of the project. The latest directory can be manually symlinked to your web root once it's uploaded (ex: `/current`), transitioning to the updated release without any delay. In the future, I plan to automatically create the symlink in SFTP connections and generally make atomic deployments much more feature rich.

To activate atomic deployments, you have to enable an `atomic` configuration option. The default directory is `releases` and that should be a good name for most people. However, if for some reason you want to change it, there's also the `releasedir` option. It must be set relative to the `path` option and must be already created on the server.

```
[production]
; ...
atomic = true
releasedir = myreleases
```

## Logging

A simple logger is available that writes on the server a `.steer-log` with information about the deployment: date and time, commit and number of changed files. By default it's disabled, but can be easily enabled by setting a `logger` configuration option:

```
[production]
; ...
logger = true
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

## Pre and Post Deployment Commands

Often it is useful to run a shell command before or after deployment. You may need to clear a cache, update some dependencies, compile assets or whatever your case is. Steer allows to run arbitrary commands either pre or post deployment over SFTP. Keep in mind that commands can't be run via FTP.

To execute commands, add one or more in the `predeploy` or `postdeploy` configuration option, separated by comma. For example:

```
[production]
;...
predeploy = rm -rf cache, touch nicefile.html
postdeploy = npm update, gulp
```

Commands are blocking, meaning that Steer will wait for them to finish before moving on to the next operation. Currently they don't produce output, but as it's trivial to add, if enough people need it, I may implement a configurable output in the future.

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

Most of the inspiration for this project comes from [PHPloy](https://github.com/banago/PHPloy), a deployment tool created by a dear friend, in which I've also been involved. I'm aiming to implement most of its features, but also have added some of my own. The main advantage of Steer right now is its performance, especially with parallel operations.