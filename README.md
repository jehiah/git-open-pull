git open-pull 
-------------

converts creates a pull request from the command line, or attaches a branch to an open github issue converting it into a pull request.

### USAGE

   $ git open-pull

### CONFIGURATION

If available, git-open-pull will use the following config values. When not available
They will be requested on the command line. Note storing your git password this way is
not secure.

    [github]
        user = ....
    [gitOpenPull]
        token = ..... # GitHub Access Token generated from https://github.com/settings/tokens
        baseAccount = ....
        baseRepo = .....
        base = master
	    # Allow maintainers of the upstream repo to modify this branch
	    # https://help.github.com/articles/allowing-changes-to-a-pull-request-branch-created-from-a-fork/
        maintainersCanModify = true | false (default: true)
    [core]
        editor = /usr/bin/vi

Hooks. git-open-pull provides the ability to modify an issue template (preProcess or postProcess) or to be notified after a PR is created (callback). pre/post process commands are executed with the first argument continaing a filename with the issue template. Callback is executed with the first argument containing the filename of a file with the json results from the GitHub api of PR details

    [gitOpenPull]
        preProcess = /path/to/exe
        postProcess = /path/to/exe
        callback = /path/to/exe

### ABOUT

Because the ideal workflow is `issue -> branch -> pull request` this script
takes a github issue and turns it into a pull request based on your branch
against master in the integration account


### Building From Source

This project uses [gb](https://getgb.io/) the Go Build tool, and has a `vendor.sh` to manage dependencies. 

```
go get github.com/constabulary/gb/...
./vendor.sh
gb build
```