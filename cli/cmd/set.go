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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var value string

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Modify current configuration",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logger.Println("No configuration key specified")
			logger.Println()
			cmd.Help()
			return
		}

		viper.Set(args[0], value)
		err := viper.SafeWriteConfig()
		if err != nil {
			err = viper.WriteConfig()
			if err != nil {
				logger.Fatal(err)
			}
		}
	},
}

func init() {
	configCmd.AddCommand(setCmd)
	setCmd.PersistentFlags().StringVar(&value, "value", "", "value to set")
}
