//
// Copyright (c) 2019-2022 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package v1

import (
	v2 "github.com/eclipse-che/che-operator/api/v2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *CheCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v2.CheCluster)
	dst.ObjectMeta = src.ObjectMeta
	return nil
}

func (dst *CheCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v2.CheCluster)
	dst.ObjectMeta = src.ObjectMeta
	return nil
}
