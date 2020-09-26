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

package v1beta1

type SecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

type RabbitConnection struct {
	Protocol           string     `json:"protocol,omitempty"`
	Host               string     `json:"host,omitempty"`
	Port               int        `json:"port,omitempty"`
	Username           string     `json:"username,omitempty"`
	PasswordSecretRef  *SecretRef `json:"passwordSecretRef,omitempty"`
	Vhost              string     `json:"vhost,omitempty"`
	InsecureSkipVerify *bool      `json:"insecureSkipVerify,omitempty"`
}
