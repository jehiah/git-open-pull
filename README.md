git open-pull 
-------------

Create a pull request from the command line, or attach a branch to an open GitHub issue converting it into a pull request.

### USAGE

   $ git open-pull


### Installing


Install from source, or visit the [releases page](https://github.com/jehiah/git-open-pull/releases)

```
go get -u github.com/jehiah/git-open-pull
````

### CONFIGURATION

If available, git-open-pull will use the following config values. When not available
They will be requested on the command line. Note: storing your GitHub API credentials this way is
not secure. Your API credentials will be stored in plain text.

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
takes a GitHub issue and turns it into a pull request based on your branch
against master in the integration account

### Building From Source

This project uses Go Modules to manage dependencies. 

```
go build
```
