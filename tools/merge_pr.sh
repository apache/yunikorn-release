#!/usr/bin/env bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

trap "control" 2

# handle interrupt and cleanup before exit
function control() {
  echo "ctrl-c caught, aborting"
  abort
}

# abort processing
function abort() {
  RESET=`git status -u -s`
  if [ -n "${RESET}" ]; then
    git reset --quiet --hard HEAD
  fi
  cleanup
  exit 1
}

# clean up temporary branches and back to checked out rev
function cleanup() {
  echo "Restoring original checkout point: ${BRANCH}"
  git checkout ${BRANCH} --quiet
  git branch -D ${PRBRANCH} --quiet
  git branch -D ${MERGEBRANCH} --quiet
  if [ ${STASHED} == "true" ]; then
		echo "  Restoring stashed files..."
  	git stash pop --quiet
  fi
}

# clean up temporary branch and stay at merge point
function leave() {
  git branch -D ${PRBRANCH} --quiet
  echo "Stopping before changed were pushed"
  echo ""
  echo "Push changes using:               git push ${REMOTE} ${MERGEBRANCH}:${MASTER}"
	echo "Revert back to original branch:   git checkout ${BRANCH}"
  echo "Cleanup of the temporary branch:  git branch -D ${MERGEBRANCH}"
  if [ ${STASHED} == "true" ]; then
  	echo "Restore stashed changes:          git stash pop"
  fi
}

# prompt for a continue response
function continue() {
  PS3=$1
  echo $2
  select yn in "yes" "no"; do
    case $REPLY in
        Y|y|yes|YES|Yes ) break;;
        N|n|no|NO|No ) return 1;;
        * ) echo "please answer: yes or no"
    esac
  done
  PS3="?#"
}

# check the jira reference 
function check_jira() {
  grep -q '^\[YUNIKORN-[0-9]\+]' <<< `echo $1`
  if [ $? -ne 0 ]; then
    echo "Subject does not contain a jira reference."
    echo "The subject line of the commit must follow the pattern:"
    echo "   [JIRA reference] subject (#PR ID)"
    echo "example: [YUNIKORN-1] Add test for commits (#1)"
    echo ""
    echo "Current subject line:"
    echo "---"
    echo "$1"
    echo "---"
    echo "Please fix the subject during the commit, press any key to continue"
    read -n 1
  fi
}

# check a temporary branch does not exist
function check_branch() {
  SHA=`git rev-parse --quiet --verify ${1}`
  if [ $? -eq 0 ]; then
    echo "branch '${1}' exists with rev '${SHA}', aborting merge"
    exit 1
  fi
}

# build the body
function create_body() {
  BODYFILE=body-pr-${PRID}-temp
  for i in `git rev-list HEAD..${PRBRANCH} --reverse`; do
    git log -1 --pretty=format:"%s%n%b" $i >> ${BODYFILE}
  done
  BODY=`tail +2 ${BODYFILE}`
  rm ${BODYFILE}
}

# input check
if [ $# -ne 1 ]; then
  NAME=`basename "$0"`
  echo "You must enter exactly 1 command line argument"
  echo "  ${NAME} PR-ID"
  echo "PR-ID: the numeric ID of the pull request, example 100"
  echo ""
  echo "Change the remote used for git commands by setting the variable REMOTE"
  echo "default is 'origin', example:"
  echo "  REMOTE=apache ${NAME} 100"
  exit 1
fi

# Allow setting the remote
REMOTE="${REMOTE:-origin}"

# merging to master branch only for now
MASTER=master
# assume a clean slate
STASHED="false"
# temporary branch IDs
PRID=$1
PRBRANCH=PR-$1
MERGEBRANCH=MERGE-PR-$1

# find git we need it
if ! command -v git &> /dev/null; then
  echo "git executable not found on path"
  exit 1
fi

# check we're in the repo, we go back to this on exit
BRANCH=`git rev-parse --abbrev-ref HEAD`
if [ $? -eq 128 ]; then
  echo "git check failed: no git repository found in the current directory"
  exit 1
fi
# save whatever we need to save before changing branches
if ! git diff-index --quiet HEAD --
then
	echo "Stashing changed and new files to create clean base"
	git stash push --include-untracked --quiet -m "before merge PR ${PRID}"
	STASHED="true"
fi

# check temp branches
check_branch ${PRBRANCH}
check_branch ${MERGEBRANCH}

# switch to a temp master
git fetch ${REMOTE} ${MASTER}:${MERGEBRANCH}
git checkout ${MERGEBRANCH} --quiet

# pull the PR down
git fetch ${REMOTE} pull/${PRID}/head:${PRBRANCH}
# check we've found the PR
if [ $? -eq 128 ]; then
  echo "github PR with ID '${PRID}' not found, aborting"
  abort
fi

# merge the PR
if ! git merge --squash ${PRBRANCH}
then
  continue "manually fix merge conflicts? " "Merge failed, conflict must be resolved before continuing"
  if [ $? -eq 1 ]; then
    echo "aborting"
    abort
  fi
  continue "continue? " "Please fix any conflicts and 'git add' conflicting files..."
  if [ $? -eq 1 ]; then
    echo "aborting"
    abort
  fi
  CONFLICT="merge conflicts resolved manually as part of commit"
fi

# Collect all the details needed for the commit:
# Assume the PR is opened by the author
# Assume the first commit has the jira reference and provides subject
# Body is concat of commit body of the first commit plus the subject and body of all follow up commits
AUTHOR=`git log HEAD..${PRBRANCH} --pretty=format:"%an <%ae>" --reverse | head -1`
SUBJECT=`git log HEAD..${PRBRANCH} --pretty=format:"%s" | tail -1`" (#${PRID})"
USER=`git config --get user.name`
EMAIL=`git config --get user.email`
SIGNED="Signed-off-by: ${USER} <${EMAIL}>"
CLOSES="Closes: #${PRID}"
create_body

# Print a message if the subject is not up to scratch
check_jira "${SUBJECT}"

# override the author email (sometimes needed)
if [ ! -z "${OVERRIDE_AUTHOR}" ]; then
  echo "override author from commit:"$'\t'${AUTHOR}
  AUTHOR=${OVERRIDE_AUTHOR}
fi

# show the collected info
echo "Commit information collected for PR: ${PRID}"
echo " author:"$'\t'${AUTHOR}
echo " subject:"$'\t'${SUBJECT}
if [ -z "${BODY}" ]; then
  echo " body:"$'\t\t'"no commit comments found"
else
  echo " body:"
  echo "------"
  echo ${BODY}
  echo "------"
fi
if [ -n "${CONFLICT}" ]; then
  echo " conflict:"$'\t'"merge conflict solved manually"
fi
echo " committer:"$'\t'${SIGNED}

if ! continue "Commit changes? " ""
then
  echo "aborting before commit"
  abort
fi

# commit the changes
if ! git commit --author "${AUTHOR}" -e -m "${SUBJECT}" -m "${BODY}" -m "${CONFLICT}" -m "${CLOSES}" -m "${SIGNED}"
then
  echo "commit failed: aborting"
  abort
fi

if ! continue "push change to ${MASTER}? " "Merge completed local ref: ${MERGEBRANCH}"
then
  if ! git push ${REMOTE} ${MERGEBRANCH}:${MASTER}
  then
    echo "Push failed"
    leave
  else
    echo "Pull request merged and pushed"
    git log -1
    cleanup
  fi
else
  echo "Exit before push"
  leave
fi
