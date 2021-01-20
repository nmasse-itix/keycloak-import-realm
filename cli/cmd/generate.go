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
	"fmt"
	"os"
	"path"

	kcimport "github.com/nmasse-itix/keycloak-realm-import"
	"github.com/spf13/cobra"
)

var realmCount, clientCount, userCount int
var targetDir string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Keycloak realms",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.MkdirAll(targetDir, 0777)
		if err != nil {
			logger.Fatal(err)
		}
		realms := kcimport.GenerateRealms(realmCount, clientCount, userCount)
		for _, realm := range realms {
			f, err := os.OpenFile(path.Join(targetDir, fmt.Sprintf("realm-%s.json", realm.ID)), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				logger.Fatal(err)
			}
			defer f.Close()
			err = kcimport.WriteRealmFile(realm, f)
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().IntVar(&realmCount, "realms", 1, "number of realms to generate")
	generateCmd.Flags().IntVar(&clientCount, "clients", 10, "number of clients to generate per realm")
	generateCmd.Flags().IntVar(&userCount, "users", 10, "number of users to generate per realm")
	generateCmd.Flags().StringVar(&targetDir, "target", ".", "target directory")
}
