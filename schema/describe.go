// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package schema

//FunctionDescription information related to a function
type FunctionDescription struct {
	Name              string
	Status            string
	Replicas          int
	AvailableReplicas int
	InvocationCount   int
	Image             string
	EnvProcess        string
	URL               string
	AsyncURL          string
	Labels            *map[string]string
	Annotations       *map[string]string
}
