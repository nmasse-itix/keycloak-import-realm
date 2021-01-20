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

package kcimport

import (
	"fmt"

	commonhttp "github.com/cloudtrust/common-service/errors"
	keycloak "github.com/cloudtrust/keycloak-client/v3"
	"github.com/cloudtrust/keycloak-client/v3/api"
)

type KeycloakCredentials struct {
	Realm    string
	Login    string
	Password string
}

type KeycloakImporter struct {
	Client      *api.Client
	Token       string
	Credentials KeycloakCredentials
}

type ImportError struct {
	StatusCode int
	Message    string
}

func (e *ImportError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func NewKeycloakImporter(config keycloak.Config) (KeycloakImporter, error) {
	var importer KeycloakImporter

	kcClient, err := api.New(config)
	if err != nil {
		return importer, err
	}

	importer.Client = kcClient

	return importer, nil
}

func (importer *KeycloakImporter) Login() error {
	token, err := importer.Client.GetToken(importer.Credentials.Realm, importer.Credentials.Login, importer.Credentials.Password)
	if err != nil {
		return err
	}

	importer.Token = token

	return nil
}

func (importer *KeycloakImporter) ApplyRealm(realm keycloak.RealmRepresentation) error {
	_, err := importer.Client.CreateRealm(importer.Token, realm)
	if err != nil {
		err := normalizeError(err)
		switch {
		case err.StatusCode == 409:
			err := importer.Client.UpdateRealm(importer.Token, *realm.ID, realm)
			if err != nil {
				err := normalizeError(err)
				return err
			}

		default:
			return err
		}
	}

	return nil
}

func (importer *KeycloakImporter) ApplyClient(realmName string, client keycloak.ClientRepresentation) error {
	if client.ClientID == nil {
		return fmt.Errorf("Missing ClientID in ClientRepresentation")
	}

	_, err := importer.Client.CreateClient(importer.Token, realmName, client)
	if err != nil {
		err := normalizeError(err)
		switch {
		case err.StatusCode == 409:
			clients, err := importer.Client.GetClients(importer.Token, realmName, "clientId", *client.ClientID)
			if err != nil {
				err := normalizeError(err)
				return err
			}

			if len(clients) != 1 {
				return fmt.Errorf("Cannot find client %s in realm %s", *client.ClientID, realmName)
			}

			existingClient := clients[0]

			err = importer.Client.UpdateClient(importer.Token, realmName, *existingClient.ID, client)
			if err != nil {
				err := normalizeError(err)
				return err
			}
		default:
			return err
		}
	}

	return nil
}

func (importer *KeycloakImporter) ApplyUser(realmName string, user keycloak.UserRepresentation) error {
	if user.Username == nil {
		return fmt.Errorf("Missing Username in UserRepresentation")
	}

	_, err := importer.Client.CreateUser(importer.Token, realmName, user)
	if err != nil {
		err := normalizeError(err)
		switch {
		case err.StatusCode == 409:
			users, err := importer.Client.GetUsers(importer.Token, realmName, "username", *user.Username)
			if err != nil {
				err := normalizeError(err)
				return err
			}

			if len(users) != 1 {
				return fmt.Errorf("Cannot find user %s in realm %s", *user.Username, realmName)
			}

			existingUser := users[0]

			err = importer.Client.UpdateUser(importer.Token, realmName, *existingUser.ID, user)
			if err != nil {
				err := normalizeError(err)
				return err
			}
		default:
			return err
		}
	}

	return nil
}

func normalizeError(err error) *ImportError {
	if e, ok := err.(commonhttp.Error); ok {
		return &ImportError{StatusCode: e.Status, Message: e.Message}
	} else if e, ok := err.(keycloak.HTTPError); ok {
		return &ImportError{StatusCode: e.HTTPStatus, Message: e.Message}
	} else if e, ok := err.(keycloak.ClientDetailedError); ok {
		return &ImportError{StatusCode: e.HTTPStatus, Message: e.Message}
	}
	return &ImportError{Message: err.Error()}
}
