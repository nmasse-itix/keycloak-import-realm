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

//go:generate statik -f -src=templates/ -include=*.template
package kcimport

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/google/uuid"
	_ "github.com/nmasse-itix/keycloak-realm-import/statik"
	"github.com/rakyll/statik/fs"
)

type GeneratedUser struct {
	ID string
}

type GeneratedClient struct {
	ID     string
	Secret string
}

type GeneratedRealm struct {
	ID      string
	Clients []GeneratedClient
	Users   []GeneratedUser
}

var defaultTemplate *template.Template
var statikFS http.FileSystem

func init() {
	var err error

	statikFS, err = fs.New()
	if err != nil {
		fmt.Printf("init: %s\n", err)
	}

	defaultTemplate, err = getTemplate(statikFS, "/realm.template")
	if err != nil {
		fmt.Printf("init: %s\n", err)
	}
}

func GenerateRealms(realmCount, clientCount, userCount int) []GeneratedRealm {
	var realms []GeneratedRealm
	for r := 0; r < realmCount; r++ {
		realm := GenerateRealm(clientCount, userCount)
		realm.ID = fmt.Sprintf("%003d", r)
		realms = append(realms, realm)
	}
	return realms
}

func GenerateRealm(clientCount, userCount int) GeneratedRealm {
	var realm GeneratedRealm
	for c := 0; c < clientCount; c++ {
		var client GeneratedClient
		client.ID = fmt.Sprintf("%006d", c)
		client.Secret = uuid.New().String()
		realm.Clients = append(realm.Clients, client)
	}
	for u := 0; u < userCount; u++ {
		var user GeneratedUser
		user.ID = fmt.Sprintf("%006d", u)
		realm.Users = append(realm.Users, user)
	}
	return realm
}

func WriteRealmFileWithTemplate(realm GeneratedRealm, out io.Writer, template *template.Template) error {
	if template == nil {
		return fmt.Errorf("No template provided")
	}

	return template.Execute(out, realm)
}
func WriteRealmFile(realm GeneratedRealm, out io.Writer) error {
	return WriteRealmFileWithTemplate(realm, out, defaultTemplate)
}

func GetRealmTemplate(content string) (*template.Template, error) {
	tmpl := template.New("realm")
	customFunctions := template.FuncMap{
		// TODO
	}
	return tmpl.Funcs(customFunctions).Parse(content)
}

func getTemplate(statikFS http.FileSystem, filename string) (*template.Template, error) {
	fd, err := statikFS.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	content, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	return GetRealmTemplate(string(content))
}
