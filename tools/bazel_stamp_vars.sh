#!/bin/bash
if [[ "${GITHUB_REF_TYPE}" =~ "tag" ]]; then
echo "VERSION ${GITHUB_REF##*/}"
echo "PHASE prod"
else
echo "VERSION $(git rev-parse HEAD)-$(date +%s)"
echo "PHASE dev"
fi
