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

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	keycloak "github.com/nmasse-itix/keycloak-client"
	kcimport "github.com/nmasse-itix/keycloak-realm-import"
	"github.com/nmasse-itix/keycloak-realm-import/async"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports realms into a Keycloak instance",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logger.Println("Nothing to import")
			logger.Println()
			cmd.Help()
			return
		}

		var realm, login, password, keycloakURL string
		realm = viper.GetString("realm")
		login = viper.GetString("login")
		password = viper.GetString("password")
		keycloakURL = viper.GetString("keycloak_url")
		missingConfig := false

		if realm == "" {
			logger.Println("Missing configuration key 'realm'")
			missingConfig = true
		}

		if login == "" {
			logger.Println("Missing configuration key 'login'")
			missingConfig = true
		}

		if password == "" {
			logger.Println("Missing configuration key 'password'")
			missingConfig = true
		}

		if keycloakURL == "" {
			logger.Println("Missing configuration key 'keycloak_url'")
			missingConfig = true
		}

		if missingConfig {
			logger.Println()
			logger.Println("Use 'kci config set' to provide the missing items.")
			logger.Fatalln()
		}

		var config keycloak.Config
		config.AddrAPI = keycloakURL
		config.AddrTokenProvider = keycloakURL + "/realms/master"
		config.Timeout = time.Duration(viper.GetInt64("http_timeout")) * time.Second

		workers := viper.GetInt("workers")
		logger.Printf("Starting import with %d workers...\n", workers)
		dispatcher, err := async.NewDispatcher(workers, config, kcimport.KeycloakCredentials{Realm: realm, Login: login, Password: password})
		if err != nil {
			logger.Fatal(err)
		}

		compileResults := make(chan struct{})
		go processResults(&dispatcher, compileResults)
		importRealms(&dispatcher, args)
		compileResults <- struct{}{}
	},
}

func importRealms(dispatcher *async.Dispatcher, files []string) {
	go dispatcher.Start()
	defer dispatcher.Stop()

	for _, file := range files {
		err := processRealmFile(file, dispatcher)
		if err != nil {
			logger.Fatal(err)
		}
	}
}

func processResults(dispatcher *async.Dispatcher, compileResults chan struct{}) {
	var count, errors, retries, oldCount int
	var empty string = ""
	var lastObject *string = &empty

	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-timer.C:
			newCount := count
			rate := newCount - oldCount
			logger.Printf("%s: %7d objects processed (%4d RPS), %7d retries, %7d errors, last object processed: %s\n", time.Now().Format("15:04:05"), newCount, rate, retries, errors, *lastObject)
			oldCount = newCount
			timer.Reset(time.Second)
		case result := <-dispatcher.Results:
			if result.Success {
				count++
				retries += result.Retries
			} else {
				errors++
				logger.Printf("%s: %s\n", result.Worker, result.Error)
			}
			lastObject = result.ObjectName()
		case <-compileResults:
			logger.Printf("%s: IMPORT IS COMPLETE. %d objects processed, %d errors\n", time.Now().Format("15:04:05"), count, errors)
			timer.Stop()
			return
		}
	}

}

func processRealmFile(filename string, dispatcher *async.Dispatcher) error {
	realmData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var realm keycloak.RealmRepresentation
	err = json.Unmarshal(realmData, &realm)
	if err != nil {
		return err
	}

	clients := realm.Clients
	users := realm.Users

	realm.Clients = &[]keycloak.ClientRepresentation{}
	realm.Users = &[]keycloak.UserRepresentation{}

	if realm.ID == nil {
		return fmt.Errorf("Missing realm ID in RealmRepresentation")
	}

	dispatcher.ApplyRealm(realm)

	if users != nil {
		for _, user := range *users {
			dispatcher.ApplyUser(*realm.ID, user)
		}
	}

	if clients != nil {
		for _, client := range *clients {
			dispatcher.ApplyClient(*realm.ID, client)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(importCmd)
	viper.SetDefault("http_timeout", 30)
	viper.SetDefault("workers", 5)
}
