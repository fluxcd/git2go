#!/usr/bin/env bash

set -euxo pipefail

tag_suffix="-nothread-check"

initial_branch="$(git rev-parse --abbrev-ref HEAD)"
repo_root="$(git rev-parse --show-toplevel)"
patch=$(cat "${repo_root}/patch/nothread.patch")

# Latest upstream tag within tag pattern.
upstream_tag="$(git ls-remote https://github.com/libgit2/git2go "refs/tags/v33.0.*" | cut --delimiter='/' --fields=3 | sort -u | tail -n1)"
if [ "" -eq "${upstream_tag}" ]; then
    echo "error finding latest upstream tag"
    exit 1
fi

new_tag="${upstream_tag}${tag_suffix}"

# Only tag if upstream has a newer tag that wasn't autopatched yet
if [ -z $(git tag -l "${new_tag}") ]; then
    # fetch upstream and create branch off latest tag.
    git remote add git2go_upstream https://github.com/libgit2/git2go
    git fetch git2go_upstream    
    git checkout -B no-thread "refs/tags/${upstream_tag}"

    # patch latest tag
    echo "${patch}" | git apply
    git add git.go
    git config --global user.email "pjbgf@linux.com"
    git config --global user.name "git2go auto-patcher"
    git commit -m "Auto-patch libgit2 nothread support"

    # tag current repo and push tag
    git tag "${new_tag}"
    git push origin "${new_tag}"

    # revert branch switch and remote creation
    git checkout "${initial_branch}"
    git remote rm git2go_upstream
else
    echo "Skipping as latest tag is already patched"
fi
