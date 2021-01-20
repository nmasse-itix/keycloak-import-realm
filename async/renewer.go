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
	"time"
)

func (tr *TokenRenewer) RenewToken(dispatcher *Dispatcher) {
	for {
		select {
		case <-tr.quit:
			return
		case <-tr.expiredToken:
			if time.Now().Sub(tr.LastTokenRenew) < 5*time.Second {
				continue
			}

			err := dispatcher.Importer.Login()
			if err != nil {
				fmt.Printf("dispatcher: Cannot renew OIDC token: %s\n", err)
				continue
			}

			tr.LastTokenRenew = time.Now()
			for i := 0; i < len(dispatcher.Workers); i++ {
				dispatcher.Workers[i].NewToken(dispatcher.Importer.Token)
			}
		}
	}
}

func (tr *TokenRenewer) Stop() {
	tr.quit <- struct{}{}
}

type TokenRenewer struct {
	quit           chan struct{}
	expiredToken   chan struct{}
	LastTokenRenew time.Time
}

func NewTokenRenewer() TokenRenewer {
	var tokenRenewer TokenRenewer
	tokenRenewer.LastTokenRenew = time.Now()
	tokenRenewer.expiredToken = make(chan struct{})
	tokenRenewer.quit = make(chan struct{})
	return tokenRenewer
}
