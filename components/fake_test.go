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
	Users  []*rabbithole.UserInfo
	Vhosts []*rabbithole.VhostInfo
	// [vhost][policyName]
	Policies map[string]map[string]*rabbithole.Policy
	// [username][vhost]
	Permissions map[string]map[string]*rabbithole.PermissionInfo
	// [vhost][queue]
	Queues map[string]map[string]*rabbithole.QueueInfo
}

var _ rabbitManager = &fakeRabbitClient{}

func newFakeRabbitClient() *fakeRabbitClient {
	return &fakeRabbitClient{
		Users:       []*rabbithole.UserInfo{},
		Vhosts:      []*rabbithole.VhostInfo{},
		Policies:    map[string]map[string]*rabbithole.Policy{},
		Permissions: map[string]map[string]*rabbithole.PermissionInfo{},
		Queues:      map[string]map[string]*rabbithole.QueueInfo{},
	}
}

func (frc *fakeRabbitClient) Factory(_uri, _user, _pass string, _t *http.Transport) (rabbitManager, error) {
	return frc, nil
}

func (frc *fakeRabbitClient) ListUsers() ([]rabbithole.UserInfo, error) {
	users := []rabbithole.UserInfo{}
	for _, user := range frc.Users {
		users = append(users, *user)
	}
	return users, nil
}

func (frc *fakeRabbitClient) GetUser(name string) (*rabbithole.UserInfo, error) {
	for _, user := range frc.Users {
		if user.Name == name {
			return user, nil
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
	frc.Users = append(frc.Users, &rabbithole.UserInfo{Name: username, PasswordHash: settings.PasswordHash, HashingAlgorithm: settings.HashingAlgorithm, Tags: settings.Tags})
	return &http.Response{StatusCode: 201}, nil
}

func (frc *fakeRabbitClient) DeleteUser(name string) (*http.Response, error) {
	for i, element := range frc.Users {
		if element.Name == name {
			copy(frc.Users[i:], frc.Users[i+1:])
			frc.Users = frc.Users[:len(frc.Users)-1]
			return &http.Response{StatusCode: 204}, nil
		}
	}
	return &http.Response{StatusCode: 404}, nil
}

func (frc *fakeRabbitClient) ListVhosts() ([]rabbithole.VhostInfo, error) {
	vhosts := []rabbithole.VhostInfo{}
	for _, vhost := range frc.Vhosts {
		vhosts = append(vhosts, *vhost)
	}
	return vhosts, nil
}

func (frc *fakeRabbitClient) GetVhost(name string) (*rabbithole.VhostInfo, error) {
	for _, vhost := range frc.Vhosts {
		if vhost.Name == name {
			return vhost, nil
		}
	}
	return nil, rabbithole.ErrorResponse{StatusCode: 404}
}

func (frc *fakeRabbitClient) PutVhost(vhost string, _settings rabbithole.VhostSettings) (*http.Response, error) {
	for _, element := range frc.Vhosts {
		if element.Name == vhost {
			return &http.Response{StatusCode: 200}, nil
		}
	}
	frc.Vhosts = append(frc.Vhosts, &rabbithole.VhostInfo{Name: vhost})
	return &http.Response{StatusCode: 201}, nil
}

func (frc *fakeRabbitClient) DeleteVhost(vhost string) (*http.Response, error) {
	for i, element := range frc.Vhosts {
		if element.Name == vhost {
			copy(frc.Vhosts[i:], frc.Vhosts[i+1:])
			frc.Vhosts = frc.Vhosts[:len(frc.Vhosts)-1]
			return &http.Response{StatusCode: 204}, nil
		}
	}
	return &http.Response{StatusCode: 404}, nil
}

func (frc *fakeRabbitClient) ListPoliciesIn(vhost string) (rec []rabbithole.Policy, err error) {
	policies := []rabbithole.Policy{}
	vhostPolicies, ok := frc.Policies[vhost]
	if ok {
		for _, policy := range vhostPolicies {
			policies = append(policies, *policy)
		}
	}
	return policies, nil
}

func (frc *fakeRabbitClient) PutPolicy(vhost string, name string, policy rabbithole.Policy) (res *http.Response, err error) {
	vhostPolicies, ok := frc.Policies[vhost]
	if !ok {
		vhostPolicies = map[string]*rabbithole.Policy{}
		frc.Policies[vhost] = vhostPolicies
	}
	_, ok = vhostPolicies[name]
	vhostPolicies[name] = &policy
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
	userPerms, ok := frc.Permissions[username]
	if !ok {
		// TODO?
		return []rabbithole.PermissionInfo{}, nil
	}
	perms := []rabbithole.PermissionInfo{}
	for _, perm := range userPerms {
		perms = append(perms, *perm)
	}
	return perms, nil
}

func (frc *fakeRabbitClient) UpdatePermissionsIn(vhost, username string, permissions rabbithole.Permissions) (res *http.Response, err error) {
	userPerms, ok := frc.Permissions[username]
	if !ok {
		userPerms = map[string]*rabbithole.PermissionInfo{}
		frc.Permissions[username] = userPerms
	}

	perm, ok := userPerms[vhost]
	if !ok {
		userPerms[vhost] = &rabbithole.PermissionInfo{
			Vhost:     vhost,
			User:      username,
			Configure: permissions.Configure,
			Read:      permissions.Read,
			Write:     permissions.Write,
		}
		return &http.Response{StatusCode: 201}, nil
	}

	perm.Read = permissions.Read
	perm.Write = permissions.Write
	perm.Configure = permissions.Configure
	return &http.Response{StatusCode: 204}, nil
}

func (frc *fakeRabbitClient) ClearPermissionsIn(vhost, username string) (res *http.Response, err error) {
	userPerms, ok := frc.Permissions[username]
	if ok {
		delete(userPerms, vhost)
		if len(userPerms) == 0 {
			delete(frc.Permissions, username)
		}
	}
	return &http.Response{StatusCode: 204}, nil
}

func (frc *fakeRabbitClient) ListQueues() ([]rabbithole.QueueInfo, error) {
	queues := []rabbithole.QueueInfo{}
	for _, vhost := range frc.Queues {
		for _, queue := range vhost {
			queues = append(queues, *queue)
		}
	}
	return queues, nil
}

func (frc *fakeRabbitClient) ListQueuesIn(vhost string) ([]rabbithole.QueueInfo, error) {
	queues := []rabbithole.QueueInfo{}
	vhostQueues, ok := frc.Queues[vhost]
	if !ok {
		// What does this actually return in real life?
		return queues, rabbithole.ErrorResponse{StatusCode: 404}
	}

	for _, queue := range vhostQueues {
		queues = append(queues, *queue)
	}

	return queues, nil
}

func (frc *fakeRabbitClient) GetQueue(vhost string, queue string) (*rabbithole.DetailedQueueInfo, error) {
	vhostQueues, ok := frc.Queues[vhost]
	if !ok {
		// What does this actually return in real life?
		return nil, rabbithole.ErrorResponse{StatusCode: 404}
	}
	queueInfo, ok := vhostQueues[queue]
	if !ok {
		// What does this actually return in real life?
		return nil, rabbithole.ErrorResponse{StatusCode: 404}
	}
	detailedInfo := &rabbithole.DetailedQueueInfo{
		Name:       queueInfo.Name,
		Vhost:      queueInfo.Vhost,
		Durable:    queueInfo.Durable,
		AutoDelete: queueInfo.AutoDelete,
		Arguments:  queueInfo.Arguments,
	}
	return detailedInfo, nil
}

func (frc *fakeRabbitClient) DeclareQueue(vhost, queue string, info rabbithole.QueueSettings) (*http.Response, error) {
	vhostQueues, ok := frc.Queues[vhost]
	if !ok {
		vhostQueues = map[string]*rabbithole.QueueInfo{}
		frc.Queues[vhost] = vhostQueues
	}
	_, ok = vhostQueues[queue]
	vhostQueues[queue] = &rabbithole.QueueInfo{
		Name:       queue,
		Vhost:      vhost,
		Durable:    info.Durable,
		AutoDelete: info.AutoDelete,
		Arguments:  info.Arguments,
	}
	if ok {
		return &http.Response{StatusCode: 200}, nil
	} else {
		return &http.Response{StatusCode: 201}, nil
	}
}

func (frc *fakeRabbitClient) DeleteQueue(vhost, queue string) (*http.Response, error) {
	vhostQueues, ok := frc.Queues[vhost]
	if !ok {
		return &http.Response{StatusCode: 404}, nil
	}
	delete(vhostQueues, queue)
	// What does this actually return in real life?
	return &http.Response{StatusCode: 204}, nil
}
