#!/usr/bin/env bash
#
# npm.sh
# @author acrazing
# @since 2017-07-19 21:15:37
# @desc npm.sh
#
set -xe

CMD="${1}"

if [ "${CMD}" == "publish" ]; then
  npm config set registry https://registry.npmjs.org/
  set +e
  npm "$@"
  npm config set registry https://registry.npm.taobao.org/
  exit $?
else:
  npm "$@"
fi
