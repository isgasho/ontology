/*
 * Copyright (C) 2019 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package message_test

import (
	"encoding/json"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	utils2 "github.com/ontio/ontology/smartcontract/service/native/utils"
)

func newTestBlockHdr() *message.ShardBlockHeader {
	hdr := &types.Header{}
	hdr.Bookkeepers = make([]keypair.PublicKey, 0)
	hdr.SigData = make([][]byte, 0)

	return &message.ShardBlockHeader{hdr}
}

func newTestShardTx(t *testing.T, version byte, shardID uint64) *message.ShardBlockTx {
	paramsBytes := []byte{1, 2, 3}
	mutable := utils.BuildNativeTransaction(utils2.ShardSysMsgContractAddress, shardsysmsg.PROCESS_CROSS_SHARD_MSG, paramsBytes)
	mutable.Version = version
	mutable.ShardID = shardID
	tx, err := mutable.IntoImmutable()
	if err != nil {
		t.Errorf("build tx failed: %s", err)
	}

	return &message.ShardBlockTx{tx}
}

func newTestShardBlockInfo(t *testing.T) *message.ShardBlockInfo {
	height := uint32(123)
	parentHeight := uint64(321)
	shardHdr := newTestBlockHdr()
	shardHdr.Header.Height = height
	shardHdr.Header.ParentHeight = parentHeight

	blkInfo := &message.ShardBlockInfo{
		FromShardID: 100,
		Height:      uint64(height),
		Header:      shardHdr,
		ShardTxs:    make(map[uint64]*message.ShardBlockTx),
	}

	version := byte(100)
	shardID := uint64(100)
	shardTx := newTestShardTx(t, version, shardID)
	blkInfo.ShardTxs[shardTx.Tx.ShardID] = shardTx

	return blkInfo
}

func TestShardBlockHeader_Marshal(t *testing.T) {
	height := uint32(123)
	parentHeight := uint64(321)

	shardHdr := newTestBlockHdr()
	shardHdr.Header.Height = height
	shardHdr.Header.ParentHeight = parentHeight

	shardBytes, err := json.Marshal(shardHdr)
	if err != nil {
		t.Fatalf("marshal shard header: %s", err)
	}

	shardHdr2 := &message.ShardBlockHeader{}
	if err := json.Unmarshal(shardBytes, shardHdr2); err != nil {
		t.Fatalf("unmarshal shard header: %s", err)
	}

	if shardHdr2.Header.ParentHeight != parentHeight {
		t.Fatalf("unmatched parent height: %d vs %d", shardHdr2.Header.ParentHeight, parentHeight)
	}

	if shardHdr2.Header.Height != height {
		t.Fatalf("unmatched height: %d vs %d", shardHdr2.Header.Height, height)
	}
}

func TestShardBlockTx_Marshal(t *testing.T) {
	version := byte(100)
	shardID := uint64(100)

	shardTx := newTestShardTx(t, version, shardID)
	shardBytes, err := json.Marshal(shardTx)
	if err != nil {
		t.Fatalf("marshal shard header: %s", err)
	}

	shardTx2 := &message.ShardBlockTx{}
	if err := json.Unmarshal(shardBytes, shardTx2); err != nil {
		t.Fatalf("unmarshal shard header: %s", err)
	}

	if shardTx2.Tx.Version != version {
		t.Fatalf("unmatched tx version: %d vs %d", shardTx2.Tx.Version, version)
	}

	if shardTx2.Tx.ShardID != shardID {
		t.Fatalf("unmatched tx type: %d vs %d", shardTx2.Tx.ShardID, shardID)
	}
}

func TestShardBlockPool(t *testing.T) {
	pool := message.NewShardBlockPool(100)
	blk := newTestShardBlockInfo(t)

	shardID := blk.FromShardID
	height := blk.Height

	if err := pool.AddBlock(blk); err != nil {
		t.Fatalf("failed add block: %s", err)
	}

	blk2 := pool.GetBlock(shardID, height)
	if blk2 == nil {
		t.Fatalf("failed get block")
	}

	if blk != blk2 {
		t.Fatalf("unmatched blk")
	}
}