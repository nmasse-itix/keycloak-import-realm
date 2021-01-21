/*
 * This file is part of the keycloak-import-realm distribution
 * (https://github.com/nmasse-itix/keycloak-import-realm).
 * Copyright (c) 2021 Nicolas Massé <nicolas.masse@itix.fr>.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package async

import (
	"fmt"

	keycloak "github.com/nmasse-itix/keycloak-client"
	kcimport "github.com/nmasse-itix/keycloak-realm-import"
)

type KeycloakUserCreationRequest struct {
	Realm string
	User  keycloak.UserRepresentation
}

type KeycloakClientCreationRequest struct {
	Realm  string
	Client keycloak.ClientRepresentation
}

type KeycloakType int

const (
	KeycloakRealm KeycloakType = iota
	KeycloakUser
	KeycloakClient
)

func (t KeycloakType) String() string {
	switch {
	case t == KeycloakClient:
		return "client"
	case t == KeycloakUser:
		return "user"
	case t == KeycloakRealm:
		return "realm"
	}

	return ""
}

type KeycloakResult struct {
	ResourceType KeycloakType
	Realm        string
	Name         string
	Success      bool
	Error        error
	Retries      int
	Worker       string
}

func NewKeycloakResult(worker string, t KeycloakType, realm *string, name *string, err error, retries int) KeycloakResult {
	res := KeycloakResult{Worker: worker, ResourceType: t}
	if err != nil {
		res.Error = err
	} else {
		res.Success = true
	}

	if realm != nil {
		res.Realm = *realm
	}

	if name != nil {
		res.Name = *name
	}

	res.Retries = retries

	return res
}

func ResultString(success bool) string {
	if success {
		return "Success"
	}

	return "Failure"
}

func (r KeycloakResult) ObjectName() *string {
	var result string
	if r.ResourceType == KeycloakRealm {
		result = fmt.Sprintf("realm %s", r.Realm)
	} else {
		result = fmt.Sprintf("%s %s/%s", r.ResourceType, r.Realm, r.Name)
	}
	return &result
}

func (r KeycloakResult) String() string {
	if r.ResourceType == KeycloakRealm {
		if r.Success {
			return fmt.Sprintf("%s => %v(type = realm, realm = %s)", r.Worker, ResultString(r.Success), r.Realm)
		}

		return fmt.Sprintf("%s => %v(type = realm, realm = %s): %s", r.Worker, ResultString(r.Success), r.Realm, r.Error)
	}

	if r.Success {
		return fmt.Sprintf("%s => %v(type = %s, realm = %s, name = %s)", r.Worker, ResultString(r.Success), r.ResourceType, r.Realm, r.Name)
	}

	return fmt.Sprintf("%s => %v(type = %s, realm = %s, name = %s): %s", r.Worker, ResultString(r.Success), r.ResourceType, r.Realm, r.Name, r.Error)
}

type Dispatcher struct {
	Client       *keycloak.Client
	Workers      []Worker
	Importer     kcimport.KeycloakImporter
	clients      chan KeycloakClientCreationRequest
	users        chan KeycloakUserCreationRequest
	Results      chan KeycloakResult
	tokenRenewer TokenRenewer
}

func NewDispatcher(workers int, config keycloak.Config, credentials kcimport.KeycloakCredentials) (Dispatcher, error) {
	var dispatcher Dispatcher

	importer, err := kcimport.NewKeycloakImporter(config)
	if err != nil {
		return dispatcher, err
	}

	importer.Credentials = credentials
	err = importer.Login()
	if err != nil {
		return dispatcher, err
	}

	dispatcher.tokenRenewer = NewTokenRenewer()
	dispatcher.Importer = importer
	dispatcher.clients = make(chan KeycloakClientCreationRequest)
	dispatcher.users = make(chan KeycloakUserCreationRequest)
	dispatcher.Results = make(chan KeycloakResult)

	dispatcher.Workers = make([]Worker, workers)
	for i := 0; i < workers; i++ {
		dispatcher.Workers[i] = NewWorker(fmt.Sprintf("worker-%03d", i), dispatcher.clients, dispatcher.users, dispatcher.Results, dispatcher.tokenRenewer.expiredToken)

		importer, err = kcimport.NewKeycloakImporter(config)
		if err != nil {
			return dispatcher, err
		}
		importer.Token = dispatcher.Importer.Token

		dispatcher.Workers[i].Importer = importer
	}

	return dispatcher, nil
}

func (dispatcher *Dispatcher) ApplyRealm(realm keycloak.RealmRepresentation) {
	var err error

	for i := 0; i < 3; i++ {
		err = dispatcher.Importer.ApplyRealm(realm)
		if err != nil {
			continue
		}
	}

	dispatcher.Results <- NewKeycloakResult("dispatcher", KeycloakRealm, realm.ID, nil, err, 0)
}

func (dispatcher *Dispatcher) ApplyClient(realmName string, client keycloak.ClientRepresentation) {
	dispatcher.clients <- KeycloakClientCreationRequest{realmName, client}
}

func (dispatcher *Dispatcher) ApplyUser(realmName string, user keycloak.UserRepresentation) {
	dispatcher.users <- KeycloakUserCreationRequest{realmName, user}
}

func (dispatcher *Dispatcher) Stop() {
	for i := 0; i < len(dispatcher.Workers); i++ {
		dispatcher.Workers[i].Stop()
	}
	dispatcher.tokenRenewer.Stop()
}

func (dispatcher *Dispatcher) Start() {
	for i := 0; i < len(dispatcher.Workers); i++ {
		go dispatcher.Workers[i].Process()
	}

	go dispatcher.tokenRenewer.RenewToken(dispatcher)

}
