/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	pi "github.com/CovenantSQL/CovenantSQL/blockproducer/interfaces"
	"github.com/CovenantSQL/CovenantSQL/proto"
)

//go:generate hsp

// SQLChainRole defines roles of account in a SQLChain.
type SQLChainRole byte

const (
	// Miner defines the miner role as a SQLChain user.
	Miner SQLChainRole = iota
	// Customer defines the customer role as a SQLChain user.
	Customer
	// NumberOfRoles defines the SQLChain roles number.
	NumberOfRoles
)

// UserPermission defines permissions of a SQLChain user.
type UserPermission int32

const (
	// Admin defines the admin user permission.
	Admin UserPermission = iota
	// Read defines the reader user permission.
	Read
	// ReadWrite defines the reader/writer user permission.
	ReadWrite
	// NumberOfUserPermission defines the user permission number.
	NumberOfUserPermission
)

// SQLChainUser defines a SQLChain user.
type SQLChainUser struct {
	Address    proto.AccountAddress
	Permission UserPermission
	Pledge uint64
}

// MinerInfo defines a miner.
type MinerInfo struct {
	address proto.AccountAddress
	activeAddress proto.AccountAddress
	name string
	PendingIncome uint64
	ReceivedIncome uint64
	Pledge uint64
}

// SQLChainProfile defines a SQLChainProfile related to an account.
type SQLChainProfile struct {
	ID      proto.DatabaseID
	Period uint64
	GasPrice uint64

	TokenBalance [SupportTokenNumber]uint64
	TokenType TokenType

	PayingBalance []uint64
	PaidBalance []uint64

	Owner   proto.AccountAddress
	Miners  []*MinerInfo

	Users   []*SQLChainUser
}

// Account store its balance, and other mate data.
type Account struct {
	Address             proto.AccountAddress
	TokenBalance		[SupportTokenNumber]uint64
	Rating              float64
	NextNonce           pi.AccountNonce
}
