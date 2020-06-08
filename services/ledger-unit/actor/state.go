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

package actor

import (
	"github.com/jancajthaml-openbank/ledger-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
)

type TransactionState struct {
	Transaction     model.Transaction
	Negotiation     map[model.Account]string
	WaitFor         map[model.Account]interface{}
	OkResponses     int
	FailedResponses int
	Ready           bool
	ReplyTo         system.Coordinates
}

func NewTransactionState() TransactionState {
	return TransactionState{
		OkResponses:     0,
		FailedResponses: 0,
		Ready:           false,
	}
}

func (state *TransactionState) Mark(response interface{}) {
	if state == nil {
		return
	}

	switch msg := response.(type) {

	case PromiseWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}

	case CommitWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}

	case RollbackWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}

	case PromiseWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	case CommitWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	case RollbackWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	case FatalErrored:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	}
}

func (state *TransactionState) ResetMarks() {
	if state == nil {
		return
	}
	state.WaitFor = make(map[model.Account]interface{})
	for account := range state.Negotiation {
		state.WaitFor[account] = nil
	}
	state.OkResponses = 0
	state.FailedResponses = 0
}

func (state TransactionState) IsNegotiationFinished() bool {
	return len(state.Negotiation) <= (state.OkResponses + state.FailedResponses)
}

func (state *TransactionState) Prepare(transaction model.Transaction, requestedBy system.Coordinates) {
	if state == nil {
		return
	}
	negotiation := transaction.PrepareRemoteNegotiation()
	state.Transaction = transaction
	state.Negotiation = negotiation
	state.ResetMarks()
	state.Ready = true
	state.ReplyTo = requestedBy
}
