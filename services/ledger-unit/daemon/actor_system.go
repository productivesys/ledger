// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/ledger-unit/config"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// ActorSystem represents actor system subroutine
type ActorSystem struct {
	system.Support
	Storage *localfs.Storage
	Metrics *Metrics
}

// NewActorSystem returns actor system fascade
func NewActorSystem(ctx context.Context, cfg config.Configuration, metrics *Metrics, storage *localfs.Storage) ActorSystem {
	return ActorSystem{
		Support: system.NewSupport(ctx, "LedgerUnit/"+cfg.Tenant, cfg.LakeHostname),
		Storage: storage,
		Metrics: metrics,
	}
}

// GreenLight daemon noop
func (system ActorSystem) GreenLight() {

}

// WaitReady wait for system to be ready
func (system ActorSystem) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-system.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}
