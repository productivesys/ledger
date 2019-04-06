// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package persistence

import (
	"strings"

	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// LoadTransaction loads transaction from journal
func LoadTransaction(storage *localfs.Storage, idTransaction string) *model.Transaction {
	transactionPath := utils.TransactionPath(idTransaction)

	data, err := storage.ReadFileFully(transactionPath)
	if err != nil {
		return nil
	}

	result := new(model.Transaction)
	result.Deserialise(data)
	return result
}

// LoadTransfer loads transfer from journal
func LoadTransfer(storage *localfs.Storage, idTransaction, idTransfer string) *model.Transfer {
	transactionPath := utils.TransactionPath(idTransaction)

	data, err := storage.ReadFileFully(transactionPath)
	if err != nil {
		return nil
	}

	result := new(model.Transaction)
	result.Deserialise(data)

	for _, transfer := range result.Transfers {
		if transfer.IDTransfer == idTransfer {
			return &transfer
		}
	}

	return nil
}

// CreateTransaction persist transaction entity state to storage
func CreateTransaction(storage *localfs.Storage) *model.Transaction {
	return PersistTransaction(storage, &model.Transaction{})
}

// StoreNewTransaction persist transaction to disk
func PersistTransaction(storage *localfs.Storage, entity *model.Transaction) *model.Transaction {
	//created := now()
	// FIXME do not store transaction like this :/ or do so for integrity?

	idTransaction := entity.IDTransaction

	transactionPath := utils.TransactionPath(idTransaction)
	transactionStatePath := utils.TransactionStatePath(idTransaction)

	if storage.WriteFile(transactionStatePath, []byte(model.StatusDirty)) != nil {
		return nil
	}

	data := entity.Serialise()
	if storage.WriteFile(transactionPath, data) != nil {
		return nil
	}

	if storage.UpdateFile(transactionStatePath, []byte(model.StatusNew)) != nil {
		return nil
	}

	return entity
}

/*
// PersistAccount persist account entity state to storage
func PersistAccount(storage *localfs.Storage, name string, entity *model.Account) *model.Account {
	data := entity.Serialise()
	path := utils.SnapshotPath(name, entity.Version)

	if storage.WriteFile(path, data) != nil {
		return nil
	}

	return entity
}*/

// GetTransactionState returns transaction state from journal
func GetTransactionState(storage *localfs.Storage, idTransaction string) (string, string) {
	fullPath := utils.TransactionStatePath(idTransaction)
	data, err := storage.ReadFileFully(fullPath)
	if err != nil {
		return "", ""
	}
	parts := strings.Split(string(data), " ")
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// IsTransferForwardedCredit returns true if transaction's credit side was forwarded
func IsTransferForwardedCredit(storage *localfs.Storage, idTransaction, idTransfer string) (bool, error) {
	fullPath := utils.TransactionForwardPath(idTransaction)
	ok, err := storage.Exists(fullPath)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	data, err := storage.ReadFileFully(fullPath)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, " ")

		if idTransfer == parts[0] && parts[1] == "credit" {
			return true, nil
		}
	}
	return false, nil
}

// IsTransferForwardedDebit returns true if transaction's debit side was forwarded
func IsTransferForwardedDebit(storage *localfs.Storage, idTransaction, idTransfer string) (bool, error) {
	fullPath := utils.TransactionForwardPath(idTransaction)
	ok, err := storage.Exists(fullPath)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	data, err := storage.ReadFileFully(fullPath)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, " ")

		if idTransfer == parts[0] && parts[1] == "debit" {
			return true, nil
		}
	}
	return false, nil
}

// AcceptForwardCredit accepts transaction credit forward request
func AcceptForwardCredit(storage *localfs.Storage, targetTenant, targetTransaction, targetTransfer, originTransaction, originTransfer string) bool {
	fullPath := utils.TransactionForwardPath(originTransaction)
	return storage.AppendFile(fullPath, []byte(originTransfer+" credit "+targetTenant+" "+targetTransaction+" "+targetTransfer)) == nil
}

// AcceptForwardDebit accepts transaction debit forward request
func AcceptForwardDebit(storage *localfs.Storage, targetTenant, targetTransaction, targetTransfer, originTransaction, originTransfer string) bool {
	fullPath := utils.TransactionForwardPath(originTransaction)
	return storage.AppendFile(fullPath, []byte(originTransfer+" debit "+targetTenant+" "+targetTransaction+" "+targetTransfer)) == nil
}

// AcceptTransaction accepts transaction
func AcceptTransaction(storage *localfs.Storage, idTransaction string) bool {
	fullPath := utils.TransactionStatePath(idTransaction)
	return storage.UpdateFile(fullPath, []byte(model.StatusAccepted)) == nil
}

// RejectTransaction rejects transaction
func RejectTransaction(storage *localfs.Storage, idTransaction string) bool {
	fullPath := utils.TransactionStatePath(idTransaction)
	return storage.UpdateFile(fullPath, []byte(model.StatusRejected)) == nil
}

// CommitTransaction changes state of transaction to committed
func CommitTransaction(storage *localfs.Storage, idTransaction string) bool {
	fullPath := utils.TransactionStatePath(idTransaction)
	return storage.UpdateFile(fullPath, []byte(model.StatusCommitted)) == nil
}

// RollbackTransaction changes state of transaction to rollbacked
func RollbackTransaction(storage *localfs.Storage, idTransaction, reason string) bool {
	fullPath := utils.TransactionStatePath(idTransaction)
	return storage.UpdateFile(fullPath, []byte(model.StatusRollbacked+" "+reason)) == nil
}
