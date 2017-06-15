git open-pull 
-------------

converts creates a pull request from the command line, or attaches a branch to an open github issue converting it into a pull request.

### USAGE

   $ git open-pull [options] [feature_branch]

### OPTIONS

   -u --user      - github username
   -p --password  - github password (not needed when passing --token, or using a saved token)
   --save         - save access token for future use
   -t --token     - github oauth token
   --otp          - one time password (two-factor auth)
   -i --issue     - issue number
   -b --base      - the branch you want your changes pulled into. default: master
   --base-account - the account containing issue and the base branch to merge into
   --base-repo    - the github repository name

   feature-branch - the branch (or git ref) where your changes are implemented
                    feature branch is assumed to be user/feature-branch if no
                    user is specified. default: working branch name (or prompted)

### CONFIGURATION

If available, git-open-pull will use the following config values. When not available
They will be requested on the command line. Note storing your git password this way is
not secure.

   [github]
           user = ....
   [gitOpenPull]
           token = .....
           baseAccount = ....
           baseRepo = .....
           base = master

### ABOUT

Because the ideal workflow is `issue -> branch -> pull request` this script
takes a github issue and turns it into a pull request based on your branch
against master in the integration account

This makes use of the github `POST /repos/:user/:repo/pulls` endpoint
more info on that is available at https://developer.github.com/v3/pulls/
