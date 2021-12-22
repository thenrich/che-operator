#!/bin/bash
#
# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation

set -e
set -x

git tag -l --sort=creatordate | tail -n 2

git remote add operator https://github.com/eclipse-che/che-operator.git
git fetch operator -q

git tag -l --sort=creatordate | tail -n 2

