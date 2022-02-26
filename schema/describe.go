// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package schema

import "github.com/openfaas/faas-provider/types"

//FunctionDescription information related to a function
type FunctionDescription struct {
	types.FunctionStatus
	Status          string
	InvocationCount int
	URL             string
	AsyncURL        string
}
