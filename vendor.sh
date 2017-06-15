#!/bin/bash

if [ -e vendor ]; then
   echo "vendor directory exists. remove before running"
   exit 1
fi


gb vendor fetch -no-recurse -revision 7a51fb928f52a196d5f31daefb8a489453ef54ff github.com/google/go-github
gb vendor fetch -no-recurse -revision f047394b6d14284165300fd82dad67edb3a4d7f6 golang.org/x/oauth2
gb vendor fetch -no-recurse -revision 53e6ce116135b80d037921a7fdd5138cf32d7a8a github.com/google/go-querystring
gb vendor fetch -no-recurse -revision ddf80d0970594e2e4cccf5a98861cad3d9eaa4cd golang.org/x/net/context
