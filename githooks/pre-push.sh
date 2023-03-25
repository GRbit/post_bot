#!/bin/bash

MESSAGING_API_ROOT_FOLDER=`echo $0 | sed 's#hooks/pre-push.sh##'`
cd $MESSAGING_API_ROOT_FOLDER || exit 2

STASH_RET=`git stash --include-untracked`

make fmt 1>/dev/null 2>/dev/null
git diff --quiet
DIFF_RET_CODE=$?

if [ $DIFF_RET_CODE -ne 0 ]
then
    echo '`make fmt` has changed your Code'
    git diff --compact-summary
    echo
    echo Looks like you forgot to run "make fmt" command on your last commit.
    echo Please fix it and try to push again
    echo

    git checkout -- .
fi

if [ "$STASH_RET" != "No local changes to save" ]
then
    git stash apply --quiet
fi

exit $DIFF_RET_CODE
