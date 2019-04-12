/*
 * Copyright (C) 2018 The ontology Authors
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

package types

import (
	"crypto/sha256"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
)

type RawHeader struct {
	Version          uint32
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	BlockRoot        common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload []byte
	NextBookkeeper   common.Address

	//Program *program.Program
	Bookkeepers [][]byte
	SigData     [][]byte

	hash *common.Uint256
}

func (bd *RawHeader) Serialization(sink *common.ZeroCopySink) error {
	_, err := bd.SerializeExt(sink)

	return err
}
func (bd *RawHeader) SerializeExt(sink *common.ZeroCopySink) (uint32, error) {
	pos := sink.Size()
	bd.serializationUnsigned(sink)
	unsignedLen := sink.Size() - pos
	sink.WriteVarUint(uint64(len(bd.Bookkeepers)))

	for _, pubkey := range bd.Bookkeepers {
		sink.WriteVarBytes(pubkey)
	}

	sink.WriteVarUint(uint64(len(bd.SigData)))
	for _, sig := range bd.SigData {
		sink.WriteVarBytes(sig)
	}

	return uint32(unsignedLen), nil
}

func (bd *RawHeader) serializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint32(bd.Version)
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteBytes(bd.BlockRoot[:])
	sink.WriteUint32(bd.Timestamp)
	sink.WriteUint32(bd.Height)
	sink.WriteUint64(bd.ConsensusData)
	sink.WriteVarBytes(bd.ConsensusPayload)
	sink.WriteBytes(bd.NextBookkeeper[:])
}

func (bd *RawHeader) Deserialization(source *common.ZeroCopySource) error {
	err := bd.deserializationUnsigned(source)
	if err != nil {
		return err
	}

	n, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	for i := 0; i < int(n); i++ {
		buf, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		bd.Bookkeepers = append(bd.Bookkeepers, buf)
	}

	m, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	for i := 0; i < int(m); i++ {
		sig, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		bd.SigData = append(bd.SigData, sig)
	}

	return nil
}

func (bd *RawHeader) deserializationUnsigned(source *common.ZeroCopySource) error {
	var irregular, eof bool

	bd.Version, eof = source.NextUint32()
	bd.PrevBlockHash, eof = source.NextHash()
	bd.TransactionsRoot, eof = source.NextHash()
	bd.BlockRoot, eof = source.NextHash()
	bd.Timestamp, eof = source.NextUint32()
	bd.Height, eof = source.NextUint32()
	bd.ConsensusData, eof = source.NextUint64()

	bd.ConsensusPayload, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	bd.NextBookkeeper, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (bd *RawHeader) Hash() common.Uint256 {
	if bd.hash != nil {
		return *bd.hash
	}
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	temp := sha256.Sum256(sink.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	bd.hash = &hash
	return hash
}

type Header struct {
	Version          uint32
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	BlockRoot        common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload []byte
	NextBookkeeper   common.Address

	//Program *program.Program
	Bookkeepers []keypair.PublicKey
	SigData     [][]byte

	hash *common.Uint256
}

func (bd *Header) Serialization(sink *common.ZeroCopySink) error {
	_, err := bd.SerializeExt(sink)

	return err
}
func (bd *Header) SerializeExt(sink *common.ZeroCopySink) (uint32, error) {
	pos := sink.Size()
	bd.serializationUnsigned(sink)
	unsignedLen := sink.Size() - pos
	sink.WriteVarUint(uint64(len(bd.Bookkeepers)))

	for _, pubkey := range bd.Bookkeepers {
		sink.WriteVarBytes(keypair.SerializePublicKey(pubkey))
	}

	sink.WriteVarUint(uint64(len(bd.SigData)))
	for _, sig := range bd.SigData {
		sink.WriteVarBytes(sig)
	}

	return uint32(unsignedLen), nil
}

//Serialize the blockheader data without program
func (bd *Header) serializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint32(bd.Version)
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteBytes(bd.BlockRoot[:])
	sink.WriteUint32(bd.Timestamp)
	sink.WriteUint32(bd.Height)
	sink.WriteUint64(bd.ConsensusData)
	sink.WriteVarBytes(bd.ConsensusPayload)
	sink.WriteBytes(bd.NextBookkeeper[:])
}

func HeaderFromRawBytes(raw []byte) (*Header, error) {
	source := common.NewZeroCopySource(raw)
	header := &Header{}
	err := header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil

}
func (bd *Header) Deserialization(source *common.ZeroCopySource) error {
	err := bd.deserializationUnsigned(source)
	if err != nil {
		return err
	}

	n, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	for i := 0; i < int(n); i++ {
		buf, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		pubkey, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		bd.Bookkeepers = append(bd.Bookkeepers, pubkey)
	}

	m, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	for i := 0; i < int(m); i++ {
		sig, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		bd.SigData = append(bd.SigData, sig)
	}

	return nil
}

func (bd *Header) deserializationUnsigned(source *common.ZeroCopySource) error {
	var irregular, eof bool

	bd.Version, eof = source.NextUint32()
	bd.PrevBlockHash, eof = source.NextHash()
	bd.TransactionsRoot, eof = source.NextHash()
	bd.BlockRoot, eof = source.NextHash()
	bd.Timestamp, eof = source.NextUint32()
	bd.Height, eof = source.NextUint32()
	bd.ConsensusData, eof = source.NextUint64()

	bd.ConsensusPayload, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	bd.NextBookkeeper, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (bd *Header) Hash() common.Uint256 {
	if bd.hash != nil {
		return *bd.hash
	}
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	temp := sha256.Sum256(sink.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	bd.hash = &hash
	return hash
}

func (bd *Header) GetMessage() []byte {
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	return sink.Bytes()
}

func (bd *Header) ToArray() []byte {
	sink := common.NewZeroCopySink(nil)
	bd.Serialization(sink)
	return sink.Bytes()
}
