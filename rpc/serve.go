/*
 * Copyright 2019 The CovenantSQL Authors.
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

package rpc

import (
	"context"
	"net"
	"net/rpc"

	"github.com/CovenantSQL/CovenantSQL/proto"
	"github.com/CovenantSQL/CovenantSQL/utils"
)

// ServeDirect serves conn directly.
func ServeDirect(ctx context.Context, server *rpc.Server, conn net.Conn, remote proto.RawNodeID) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	nodeAwareCodec := NewNodeAwareServerCodec(ctx, utils.GetMsgPackServerCodec(conn), &remote)
	server.ServeCodec(nodeAwareCodec)
}
