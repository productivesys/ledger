// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"github.com/jancajthaml-openbank/ledger-rest/actor"
	"github.com/jancajthaml-openbank/ledger-rest/model"
	"github.com/jancajthaml-openbank/ledger-rest/persistence"
	localfs "github.com/jancajthaml-openbank/local-fs"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

// GetTransaction returns transaction state
func GetTransaction(storage localfs.Storage) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		id := c.Param("id")
		if id == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}

		transaction, err := persistence.LoadTransaction(storage, tenant, id)
		if err != nil {
			return err
		}

		if transaction == nil {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}

		chunk, err := json.Marshal(transaction)
		if err != nil {
			return err
		}

		c.Response().WriteHeader(http.StatusOK)
		c.Response().Write(chunk)
		c.Response().Flush()
		return nil
	}
}

// CreateTransaction creates new transaction for given tenant
func CreateTransaction(storage localfs.Storage, system *actor.System) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}

		b, err := ioutil.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return err
		}

		var req = new(model.Transaction)
		if json.Unmarshal(b, req) != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return nil
		}

		switch actor.CreateTransaction(system, tenant, *req).(type) {

		case *actor.TransactionCreated:
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
			c.Response().WriteHeader(http.StatusOK)
			c.Response().Write([]byte(req.IDTransaction))
			c.Response().Flush()
			return nil

		case *actor.TransactionRejected:
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
			c.Response().WriteHeader(http.StatusCreated)
			c.Response().Write([]byte(req.IDTransaction))
			c.Response().Flush()
			return nil

		case *actor.TransactionRefused:
			c.Response().WriteHeader(http.StatusExpectationFailed)
			return nil

		case *actor.TransactionDuplicate:
			c.Response().WriteHeader(http.StatusConflict)
			return nil

		case *actor.TransactionRace, *actor.ReplyTimeout:
			c.Response().WriteHeader(http.StatusAccepted)
			return nil

		default:
			return err

		}
	}
}

// GetTransactions return existing transactions of given tenant
func GetTransactions(storage localfs.Storage) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			c.Response().WriteHeader(http.StatusNotFound)
			return nil
		}

		transactions, err := persistence.LoadTransactionsIDs(storage, tenant)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)

		for idx, transaction := range transactions {
			if idx == len(transactions)-1 {
				c.Response().Write([]byte(transaction))
			} else {
				c.Response().Write([]byte(transaction + "\n"))
			}
			c.Response().Flush()
		}

		return nil
	}
}
