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
	kcimport "github.com/nmasse-itix/keycloak-realm-import"
)

type Worker struct {
	clients      chan KeycloakClientCreationRequest
	users        chan KeycloakUserCreationRequest
	quit         chan struct{}
	results      chan KeycloakResult
	Importer     kcimport.KeycloakImporter
	newToken     chan string
	Identity     string
	expiredToken chan struct{}
}

func NewWorker(identity string, clients chan KeycloakClientCreationRequest, users chan KeycloakUserCreationRequest, results chan KeycloakResult, expiredToken chan struct{}) Worker {
	var worker Worker
	worker.clients = clients
	worker.quit = make(chan struct{})
	worker.results = results
	worker.users = users
	worker.newToken = make(chan string, 1)
	worker.Identity = identity
	worker.expiredToken = expiredToken
	return worker
}

func (worker *Worker) Process() {
	for {
		select {
		case newToken := <-worker.newToken:
			worker.Importer.Token = newToken
		case request := <-worker.users:
			var err error
			var retries int
			for retries = 0; retries < 3; retries++ {
				err = worker.Importer.ApplyUser(request.Realm, request.User)
				if err == nil {
					break
				}

				if e, ok := err.(*kcimport.ImportError); ok {
					if e.StatusCode == 401 {
						worker.expiredToken <- struct{}{}
						select {
						case newToken := <-worker.newToken:
							worker.Importer.Token = newToken
							continue
						}
					}
				}
			}
			worker.results <- NewKeycloakResult(worker.Identity, KeycloakUser, &request.Realm, request.User.Username, err, retries)
		case request := <-worker.clients:
			var err error
			var retries int
			for retries = 0; retries < 3; retries++ {
				err = worker.Importer.ApplyClient(request.Realm, request.Client)
				if err == nil {
					break
				}

				if e, ok := err.(*kcimport.ImportError); ok {
					if e.StatusCode == 401 {
						worker.expiredToken <- struct{}{}
						select {
						case newToken := <-worker.newToken:
							worker.Importer.Token = newToken
							continue
						}
					}
				}

			}
			worker.results <- NewKeycloakResult(worker.Identity, KeycloakClient, &request.Realm, request.Client.ClientID, err, retries)
		case <-worker.quit:
			return
		}
	}
}

func (worker *Worker) NewToken(token string) {
	worker.newToken <- token
}

func (worker *Worker) Stop() {
	worker.quit <- struct{}{}
}
