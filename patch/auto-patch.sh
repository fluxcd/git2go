#!/usr/bin/env bash

set -euxo pipefail

tag_suffix="-flux"

initial_branch="$(git rev-parse --abbrev-ref HEAD)"
repo_root="$(git rev-parse --show-toplevel)"

# Latest upstream tag within tag pattern.
upstream_tag=$(git ls-remote https://github.com/libgit2/git2go "refs/tags/v33.0.*" | cut --delimiter='/' --fields=3 | sort -u | tail -n1)
if [ -z "${upstream_tag}" ]; then
    echo "error finding latest upstream tag"
    exit 1
fi

new_tag="${upstream_tag}${tag_suffix}"

# Only tag if upstream has a newer tag that wasn't autopatched yet.
if [ -z $(git tag -l "${new_tag}") ]; then
    # Fetch upstream and create branch off latest tag.
    git remote add git2go_upstream https://github.com/libgit2/git2go
    git fetch git2go_upstream

    git config --global user.email "fluxcdbot@users.noreply.github.com"
    git config --global user.name "fluxcdbot"
    
    # Create a temporary dir to store patches before checking out
    # upstream tag. Then ensures said dir is removed.
    TMPDIR="$(mktemp -d /tmp/auto-patch_XXXXXX)"
    trap "rm -rf ${TMPDIR}" EXIT
    cp "${repo_root}"/patch/*.patch "${TMPDIR}"

    git checkout -B auto-patch "refs/tags/${upstream_tag}"

    # Apply patch files.
    for PATCH_FILE in "${TMPDIR}"/*.patch; do
        git apply < "${PATCH_FILE}"
        git add --all
        git commit -m "Apply $(basename ${PATCH_FILE})"
    done

    # Tag current repo and push tag.
    git tag "${new_tag}"
    git push origin "${new_tag}"

    # Revert branch switch and remove temporary remote.
    git checkout "${initial_branch}"
    git remote rm git2go_upstream
else
    echo "Skipping as ${new_tag} already exists"
fi
