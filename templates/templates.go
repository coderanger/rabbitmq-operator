// +build !release

/*
Copyright 2020 Noah Kantrowitz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

import (
	"net/http"
	"path"
	"runtime"

	templateutils "github.com/coderanger/controller-utils/templates"
)

//go:generate go run ../hack/assets_generate.go
var Templates http.FileSystem

func init() {
	_, line, _, ok := runtime.Caller(0)
	if !ok {
		panic("Unable to find caller line")
	}
	Templates = templateutils.NewFilteredFileSystem(http.Dir(path.Dir(line))).Exclude("*.go")
}
