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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CovenantSQL/CovenantSQL/kayak"
	"github.com/CovenantSQL/CovenantSQL/proto"
	"github.com/CovenantSQL/CovenantSQL/route"
	"github.com/CovenantSQL/CovenantSQL/rpc"
)

// copy from /sqlchain/runtime.go
// rt define the runtime of main chain.
type rt struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	// chainInitTime is the initial cycle time, when the Genesis block is produced.
	chainInitTime time.Time

	accountAddress proto.AccountAddress
	server         *rpc.Server

	bpNum uint32
	// index is the index of the current server in the peer list.
	index       uint32
	leaderIndex uint32

	// period is the block producing cycle.
	period time.Duration
	// tick defines the maximum duration between each cycle.
	tick        time.Duration
	updateTerms uint64

	// peersMutex protects following peers-relative fields.
	peersMutex sync.Mutex
	peers      *proto.Peers
	nodeID     proto.NodeID

	stateMutex sync.Mutex // Protects following fields.
	// nextTurn is the height of the next block.
	nextTurn uint32
	// head is the current head of the best chain.
	head *State

	// timeMutex protects following time-relative fields.
	timeMutex sync.Mutex
	// offset is the time difference calculated by: coodinatedChainTime - time.Now().
	//
	// TODO(leventeliu): update offset in ping cycle.
	offset time.Duration
}

// now returns the current coodinated chain time.
func (r *rt) now() time.Time {
	r.timeMutex.Lock()
	defer r.timeMutex.Unlock()
	// TODO(lambda): why does sqlchain not need UTC
	return time.Now().UTC().Add(r.offset)
}

func newRuntime(ctx context.Context, cfg *Config, accountAddress proto.AccountAddress) *rt {
	var (
		index, lindex int32
		cld, ccl      = context.WithCancel(ctx)
	)
	index, _ = cfg.Peers.Find(cfg.NodeID)
	lindex, _ = cfg.Peers.Find(cfg.Peers.Leader)
	return &rt{
		ctx:            cld,
		cancel:         ccl,
		wg:             &sync.WaitGroup{},
		chainInitTime:  cfg.Genesis.SignedHeader.Timestamp,
		accountAddress: accountAddress,
		server:         cfg.Server,
		bpNum:          uint32(len(cfg.Peers.Servers)),
		index:          uint32(index),
		leaderIndex:    uint32(lindex),
		period:         cfg.Period,
		tick:           cfg.Tick,
		updateTerms:    cfg.UpdateTerms,
		peers:          cfg.Peers,
		nodeID:         cfg.NodeID,
		nextTurn:       1,
		head:           &State{},
		offset:         time.Duration(0),
	}
}

func (r *rt) startService(chain *Chain) {
	r.server.RegisterService(route.BlockProducerRPCName, &ChainRPCService{chain: chain})
	r.server.RegisterService("Kayak", &chain.ka)
}

// nextTick returns the current clock reading and the duration till the next turn. If duration
// is less or equal to 0, use the clock reading to run the next cycle - this avoids some problem
// caused by concurrently time synchronization.
func (r *rt) nextTick() (t time.Time, d time.Duration) {
	t = r.now()
	d = r.chainInitTime.Add(time.Duration(r.nextTurn) * r.period).Sub(t)

	if d > r.tick {
		d = r.tick
	}

	return
}

func (r *rt) isMyTurn() bool {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	return r.nextTurn%r.bpNum == r.index
}

// setNextTurn prepares the runtime state for the next turn.
func (r *rt) setNextTurn() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	r.nextTurn++
}

func (r *rt) getIndexTotalServer() (uint32, uint32, proto.NodeID) {
	return r.index, r.bpNum, r.nodeID
}

func (r *rt) getPeerInfoString() string {
	index, bpNum, nodeID := r.getIndexTotalServer()
	return fmt.Sprintf("[%d/%d] %s", index, bpNum, nodeID)
}

// getHeightFromTime calculates the height with this sql-chain config of a given time reading.
func (r *rt) getHeightFromTime(t time.Time) uint32 {
	return uint32(t.Sub(r.chainInitTime) / r.period)
}

func (r *rt) getNextTurn() uint32 {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()

	return r.nextTurn
}

func (r *rt) getPeers() *proto.Peers {
	r.peersMutex.Lock()
	defer r.peersMutex.Unlock()
	peers := r.peers.Clone()
	return &peers
}

func (r *rt) rotateLeader(count uint64, ka *kayak.Runtime) {
	if r.updateTerms <= 0 {
		return
	}
	r.peersMutex.Lock()
	defer r.peersMutex.Unlock()
	var term = count / r.updateTerms
	r.peers = &proto.Peers{
		PeersHeader: proto.PeersHeader{
			Version: r.peers.Version,
			Term:    term,
			Leader:  r.peers.Servers[(uint64(r.leaderIndex)+term)%uint64(r.bpNum)],
			Servers: r.peers.Servers,
		},
	}
	ka.UpdatePeers(r.peers)
}

func (r *rt) getHead() *State {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	return r.head
}

func (r *rt) setHead(head *State) {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	r.head = head
}

func (r *rt) stop() {
	r.cancel()
	r.wg.Wait()
}

func (r *rt) goFunc(f func(ctx context.Context)) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		f(r.ctx)
	}()
}
