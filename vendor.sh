#!/bin/bash

if [ -e vendor ]; then
   echo "vendor directory exists. remove before running"
   exit 1
fi


gb vendor fetch -no-recurse -revision 7a51fb928f52a196d5f31daefb8a489453ef54ff github.com/google/go-github
