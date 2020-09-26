/*
Copyright 2020 Noah Kantrowitz
Copyright 2019 Ridecell, Inc.

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

package components

import (
	"net/http"

	rabbithole "github.com/michaelklishin/rabbit-hole"
)

type fakeRabbitClient struct {
	Users       []rabbithole.UserInfo
	Vhosts      []rabbithole.VhostInfo
	Policies    map[string]map[string]rabbithole.Policy
	Permissions map[string][]rabbithole.PermissionInfo
}

func newFakeRabbitClient() *fakeRabbitClient {
	return &fakeRabbitClient{
		Users:       []rabbithole.UserInfo{},
		Vhosts:      []rabbithole.VhostInfo{},
		Policies:    make(map[string]map[string]rabbithole.Policy),
		Permissions: make(map[string][]rabbithole.PermissionInfo),
	}
}

func (frc *fakeRabbitClient) Factory(_uri, _user, _pass string, _t *http.Transport) (rabbitManager, error) {
	return frc, nil
}

func (frc *fakeRabbitClient) ListUsers() ([]rabbithole.UserInfo, error) {
	return frc.Users, nil
}

func (frc *fakeRabbitClient) GetUser(name string) (*rabbithole.UserInfo, error) {
	for _, user := range frc.Users {
		if user.Name == name {
			return &user, nil
		}
	}
	return nil, rabbithole.ErrorResponse{StatusCode: 404}
}

func (frc *fakeRabbitClient) PutUser(username string, settings rabbithole.UserSettings) (*http.Response, error) {
	for _, user := range frc.Users {
		if user.Name == username {
			user.PasswordHash = settings.PasswordHash
			user.HashingAlgorithm = settings.HashingAlgorithm
			user.Tags = settings.Tags
			return &http.Response{StatusCode: 200}, nil
		}
	}
	frc.Users = append(frc.Users, rabbithole.UserInfo{Name: username, PasswordHash: settings.Password, Tags: settings.Tags})
	return &http.Response{StatusCode: 201}, nil
}

func (frc *fakeRabbitClient) ListVhosts() ([]rabbithole.VhostInfo, error) {
	return frc.Vhosts, nil
}

func (frc *fakeRabbitClient) PutVhost(vhost string, _settings rabbithole.VhostSettings) (*http.Response, error) {
	for _, element := range frc.Vhosts {
		if element.Name == vhost {
			return &http.Response{StatusCode: 200}, nil
		}
	}
	frc.Vhosts = append(frc.Vhosts, rabbithole.VhostInfo{Name: vhost})
	return &http.Response{StatusCode: 201}, nil
}

func (frc *fakeRabbitClient) ListPoliciesIn(vhost string) (rec []rabbithole.Policy, err error) {
	policies := []rabbithole.Policy{}
	vhostPolicies, ok := frc.Policies[vhost]
	if ok {
		for _, policy := range vhostPolicies {
			policies = append(policies, policy)
		}
	}
	return policies, nil
}

func (frc *fakeRabbitClient) PutPolicy(vhost string, name string, policy rabbithole.Policy) (res *http.Response, err error) {
	vhostPolicies, ok := frc.Policies[vhost]
	if !ok {
		vhostPolicies = map[string]rabbithole.Policy{}
		frc.Policies[vhost] = vhostPolicies
	}
	_, ok = vhostPolicies[name]
	vhostPolicies[name] = policy
	if ok {
		return &http.Response{StatusCode: 200}, nil
	} else {
		return &http.Response{StatusCode: 201}, nil

	}
}

func (frc *fakeRabbitClient) DeletePolicy(vhost string, name string) (res *http.Response, err error) {
	vhostPolicies, ok := frc.Policies[vhost]
	if !ok {
		return &http.Response{StatusCode: 404}, nil
	}
	delete(vhostPolicies, name)
	// If the policy exists or not both api calls returns 204
	return &http.Response{StatusCode: 204}, nil
}

func (frc *fakeRabbitClient) ListPermissionsOf(username string) (rec []rabbithole.PermissionInfo, err error) {
	return frc.Permissions[username], nil
}
func (frc *fakeRabbitClient) UpdatePermissionsIn(vhost, username string, permissions rabbithole.Permissions) (res *http.Response, err error) {
	for key := range frc.Permissions[username] {
		//Update
		if frc.Permissions[username][key].Vhost == vhost {
			frc.Permissions[username][key].Configure = permissions.Configure
			frc.Permissions[username][key].Read = permissions.Read
			frc.Permissions[username][key].Write = permissions.Write
			return &http.Response{StatusCode: 204}, nil
		}
	}
	//Create
	frc.Permissions[username] = append(frc.Permissions[username], rabbithole.PermissionInfo{
		Vhost:     vhost,
		User:      username,
		Configure: permissions.Configure,
		Read:      permissions.Read,
		Write:     permissions.Write,
	})
	return &http.Response{StatusCode: 201}, nil
}
func (frc *fakeRabbitClient) ClearPermissionsIn(vhost, username string) (res *http.Response, err error) {
	for key := range frc.Permissions[username] {
		if frc.Permissions[username][key].Vhost == vhost {
			frc.Permissions[username][key] = frc.Permissions[username][len(frc.Permissions[username])-1]
			frc.Permissions[username] = frc.Permissions[username][:len(frc.Permissions[username])-1]
			return &http.Response{StatusCode: 204}, nil
		}
	}
	return &http.Response{StatusCode: 204}, nil
}
