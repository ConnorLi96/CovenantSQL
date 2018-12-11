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

package blockproducer

import (
	"bytes"
	"database/sql"

	pi "github.com/CovenantSQL/CovenantSQL/blockproducer/interfaces"
	"github.com/CovenantSQL/CovenantSQL/crypto/hash"
	"github.com/CovenantSQL/CovenantSQL/proto"
	"github.com/CovenantSQL/CovenantSQL/types"
	"github.com/CovenantSQL/CovenantSQL/utils"
	"github.com/CovenantSQL/CovenantSQL/utils/log"
	xi "github.com/CovenantSQL/CovenantSQL/xenomint/interfaces"
	xs "github.com/CovenantSQL/CovenantSQL/xenomint/sqlite"
)

var (
	ddls = [...]string{
		// Chain state tables
		`CREATE TABLE IF NOT EXISTS "blocks" (
	"height"    INT
	"hash"		TEXT
	"parent"	TEXT
	"encoded"	BLOB
	UNIQUE INDEX ("hash")
)`,
		`CREATE TABLE IF NOT EXISTS "txPool" (
	"type"		INT
	"hash"		TEXT
	"encoded"	BLOB
	UNIQUE INDEX ("hash")
)`,
		`CREATE TABLE IF NOT EXISTS "irreversible" (
	"id"		INT
	"hash"		TEXT
)`,
		// Meta state tables
		`CREATE TABLE IF NOT EXISTS "accounts" (
	"address"	TEXT
	"encoded"	BLOB
	UNIQUE INDEX ("address")
)`,
		`CREATE TABLE IF NOT EXISTS "shardChain" (
	"address"	TEXT
	"id"		TEXT
	"encoded"	BLOB
	UNIQUE INDEX ("address", "id")
)`,
		`CREATE TABLE IF NOT EXISTS "provider" (
	"address"	TEXT
	"encoded"	BLOB
	UNIQUE INDEX ("address")
)`,
	}
)

type storageProcedure func(tx *sql.Tx) error
type storageCallback func()

func store(st xi.Storage, sps []storageProcedure, cb storageCallback) (err error) {
	var tx *sql.Tx
	// BEGIN
	if tx, err = st.Writer().Begin(); err != nil {
		return
	}
	// ROLLBACK on failure
	defer tx.Rollback()
	// WRITE
	for _, sp := range sps {
		if err = sp(tx); err != nil {
			return
		}
	}
	// CALLBACK: MUST NOT FAIL
	cb()
	// COMMIT
	if err = tx.Commit(); err != nil {
		log.WithError(err).Fatalf("Failed to commit storage transaction")
	}
	return
}

func errPass(err error) storageProcedure {
	return func(_ *sql.Tx) error {
		return err
	}
}

func openStorage(path string) (st xi.Storage, err error) {
	if st, err = xs.NewSqlite(path); err != nil {
		return
	}
	for _, v := range ddls {
		if _, err = st.Writer().Exec(v); err != nil {
			return
		}
	}
	return
}

func addBlock(height uint32, b *types.BPBlock) storageProcedure {
	var (
		enc *bytes.Buffer
		err error
	)
	if enc, err = utils.EncodeMsgPack(b); err != nil {
		return errPass(err)
	}
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`INSERT OR REPLACE INTO "blocks" VALUES (?, ?, ?, ?)`,
			height,
			b.BlockHash().String(),
			b.ParentHash().String(),
			enc.Bytes())
		return
	}
}

func addTx(t pi.Transaction) storageProcedure {
	var (
		tt  = t
		enc *bytes.Buffer
		err error
	)
	if _, ok := tt.(*pi.TransactionWrapper); !ok {
		tt = pi.WrapTransaction(tt)
	}
	if enc, err = utils.EncodeMsgPack(tt); err != nil {
		return errPass(err)
	}
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`INSERT OR REPLACE INTO "txPool" VALUES (?, ?, ?)`,
			uint32(t.GetTransactionType()),
			t.Hash().String(),
			enc.Bytes())
		return
	}
}

func updateIrreversible(h hash.Hash) storageProcedure {
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`INSERT OR REPLACE INTO "irreversible" VALUES (?, ?)`, 0, h.String())
		return
	}
}

func deleteTxs(txs []pi.Transaction) storageProcedure {
	var hs = make([]hash.Hash, len(txs))
	for i, v := range txs {
		hs[i] = v.Hash()
	}
	return func(tx *sql.Tx) (err error) {
		var stmt *sql.Stmt
		if stmt, err = tx.Prepare(`DELETE FROM "txPool" WHERE "hash"=?`); err != nil {
			return
		}
		defer stmt.Close()
		for _, v := range hs {
			if _, err = stmt.Exec(v.String()); err != nil {
				return
			}
		}
		return
	}
}

func updateAccount(account *types.Account) storageProcedure {
	var (
		enc *bytes.Buffer
		err error
	)
	if enc, err = utils.EncodeMsgPack(account); err != nil {
		return errPass(err)
	}
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`INSERT OR REPLACE INTO "accounts" VALUES (?, ?)`,
			account.Address.String(),
			enc.Bytes())
		return
	}
}

func deleteAccount(address proto.AccountAddress) storageProcedure {
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`DELETE FROM "accounts" WHERE "address"=?`, address.String())
		return
	}
}

func updateShardChain(profile *types.SQLChainProfile) storageProcedure {
	var (
		enc *bytes.Buffer
		err error
	)
	if enc, err = utils.EncodeMsgPack(profile); err != nil {
		return errPass(err)
	}
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`INSERT OR REPLACE INTO "shardChain" VALUES (?, ?, ?)`,
			profile.Address.String(),
			string(profile.ID),
			enc.Bytes())
		return
	}
}

func deleteShardChain(id proto.DatabaseID) storageProcedure {
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`DELETE FROM "shardChain" WHERE "id"=?`, id)
		return
	}
}

func updateProvider(profile *types.ProviderProfile) storageProcedure {
	var (
		enc *bytes.Buffer
		err error
	)
	if enc, err = utils.EncodeMsgPack(profile); err != nil {
		return errPass(err)
	}
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`INSERT OR REPLACE INTO "provider" VALUES (?, ?)`,
			profile.Provider.String(),
			enc.Bytes())
		return
	}
}

func deleteProvider(address proto.AccountAddress) storageProcedure {
	return func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`DELETE FROM "provider" WHERE "address"=?`, address.String())
		return
	}
}

func loadIrreHash(st xi.Storage) (irre hash.Hash, err error) {
	var hex string
	// Load last irreversible block hash
	if err = st.Reader().QueryRow(
		`SELECT "hash" FROM "irreversible" WHERE "id"=0`,
	).Scan(&hex); err != nil {
		return
	}
	if err = hash.Decode(&irre, hex); err != nil {
		return
	}
	return
}

func loadTxPool(st xi.Storage) (txPool map[hash.Hash]pi.Transaction, err error) {
	var (
		th   hash.Hash
		rows *sql.Rows
		tt   uint32
		hex  string
		enc  []byte
		pool = make(map[hash.Hash]pi.Transaction)
	)

	if rows, err = st.Reader().Query(
		`SELECT "type", "hash", "encoded" FROM "txPool"`,
	); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&tt, &hex, &enc); err != nil {
			return
		}
		if err = hash.Decode(&th, hex); err != nil {
			return
		}
		var dec = &pi.TransactionWrapper{}
		if err = utils.DecodeMsgPack(enc, dec); err != nil {
			return
		}
		pool[th] = dec.Unwrap()
	}

	txPool = pool
	return
}

func loadBlocks(
	st xi.Storage, irreHash hash.Hash) (irre *blockNode, heads []*blockNode, err error,
) {
	var (
		rows *sql.Rows

		genesis    = hash.Hash{}
		index      = make(map[hash.Hash]*blockNode)
		headsIndex = make(map[hash.Hash]*blockNode)

		// Scan buffer
		v1     uint32
		v2, v3 string
		v4     []byte

		ok     bool
		bh, ph hash.Hash
		bn, pn *blockNode
	)

	// Load blocks
	if rows, err = st.Reader().Query(
		`SELECT "height", "hash", "parent", "encoded" FROM "blocks" ORDER BY "rowid"`,
	); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Scan and decode block
		if err = rows.Scan(&v1, &v2, &v3, &v4); err != nil {
			return
		}
		if err = hash.Decode(&bh, v2); err != nil {
			return
		}
		if err = hash.Decode(&ph, v3); err != nil {
			return
		}
		var dec = &types.BPBlock{}
		if err = utils.DecodeMsgPack(v4, dec); err != nil {
			return
		}
		// Add genesis block
		if ph.IsEqual(&genesis) {
			if _, ok = index[ph]; ok {
				err = ErrMultipleGenesis
				return
			}
			bn = newBlockNodeEx(0, dec, nil)
			index[bh] = bn
			headsIndex[bh] = bn
			return
		}
		// Add normal block
		if pn, ok = index[ph]; ok {
			err = ErrParentNotFound
			return
		}
		bn = newBlockNodeEx(v1, dec, pn)
		index[bh] = bn
		if _, ok = headsIndex[ph]; ok {
			delete(headsIndex, ph)
		}
		headsIndex[bh] = bn
	}

	if irre, ok = index[irreHash]; !ok {
		err = ErrParentNotFound
		return
	}

	for _, v := range headsIndex {
		heads = append(heads, v)
	}

	return
}

func loadAndCacheAccounts(st xi.Storage, view *metaState) (err error) {
	var (
		rows *sql.Rows
		hex  string
		addr hash.Hash
		enc  []byte
	)

	if rows, err = st.Reader().Query(`SELECT "address", "encoded" FROM "accounts"`); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&hex, &enc); err != nil {
			return
		}
		if err = hash.Decode(&addr, hex); err != nil {
			return
		}
		var dec = &accountObject{}
		if err = utils.DecodeMsgPack(enc, &dec.Account); err != nil {
			return
		}
		view.readonly.accounts[proto.AccountAddress(addr)] = dec
	}

	return
}

func loadAndCacheShardChainProfiles(st xi.Storage, view *metaState) (err error) {
	var (
		rows *sql.Rows
		id   string
		enc  []byte
	)

	if rows, err = st.Reader().Query(`SELECT "id", "encoded" FROM "shardChain"`); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&id, &enc); err != nil {
			return
		}
		var dec = &sqlchainObject{}
		if err = utils.DecodeMsgPack(enc, &dec.SQLChainProfile); err != nil {
			return
		}
		view.readonly.databases[proto.DatabaseID(id)] = dec
	}

	return
}

func loadAndCacheProviders(st xi.Storage, view *metaState) (err error) {
	var (
		rows *sql.Rows
		hex  string
		addr hash.Hash
		enc  []byte
	)

	if rows, err = st.Reader().Query(`SELECT "address", "encoded" FROM "provider"`); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&hex, &enc); err != nil {
			return
		}
		if err = hash.Decode(&addr, hex); err != nil {
			return
		}
		var dec = &providerObject{}
		if err = utils.DecodeMsgPack(enc, &dec.ProviderProfile); err != nil {
			return
		}
		view.readonly.provider[proto.AccountAddress(addr)] = dec
	}

	return
}

func loadImmutableState(st xi.Storage) (immutable *metaState, err error) {
	immutable = newMetaState()
	if err = loadAndCacheAccounts(st, immutable); err != nil {
		return
	}
	if err = loadAndCacheShardChainProfiles(st, immutable); err != nil {
		return
	}
	if err = loadAndCacheProviders(st, immutable); err != nil {
		return
	}
	return
}

func loadDatabase(st xi.Storage) (
	irre *blockNode,
	heads []*blockNode,
	immutable *metaState,
	txPool map[hash.Hash]pi.Transaction,
	err error,
) {
	var irreHash hash.Hash
	// Load last irreversible block hash
	if irreHash, err = loadIrreHash(st); err != nil {
		return
	}
	// Load blocks
	if irre, heads, err = loadBlocks(st, irreHash); err != nil {
		return
	}
	// Load immutable state
	if immutable, err = loadImmutableState(st); err != nil {
		return
	}
	// Load tx pool
	if txPool, err = loadTxPool(st); err != nil {
		return
	}

	return
}