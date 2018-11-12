package types

// Code generated by github.com/CovenantSQL/HashStablePack DO NOT EDIT.

import (
	hsp "github.com/CovenantSQL/HashStablePack/marshalhash"
)

// MarshalHash marshals for hash
func (z *Account) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 6
	o = append(o, 0x86, 0x86)
	o = hsp.AppendArrayHeader(o, uint32(SupportTokenNumber))
	for za0001 := range z.TokenBalance {
		o = hsp.AppendUint64(o, z.TokenBalance[za0001])
	}
	o = append(o, 0x86)
	o = hsp.AppendFloat64(o, z.Rating)
	o = append(o, 0x86)
	if oTemp, err := z.NextNonce.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	o = append(o, 0x86)
	if oTemp, err := z.Address.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	o = append(o, 0x86)
	o = hsp.AppendUint64(o, z.StableCoinBalance)
	o = append(o, 0x86)
	o = hsp.AppendUint64(o, z.CovenantCoinBalance)
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Account) Msgsize() (s int) {
	s = 1 + 13 + hsp.ArrayHeaderSize + (int(SupportTokenNumber) * (hsp.Uint64Size)) + 7 + hsp.Float64Size + 10 + z.NextNonce.Msgsize() + 8 + z.Address.Msgsize() + 18 + hsp.Uint64Size + 20 + hsp.Uint64Size
	return
}

// MarshalHash marshals for hash
func (z *SQLChainProfile) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 5
	o = append(o, 0x85, 0x85)
	o = hsp.AppendArrayHeader(o, uint32(len(z.Users)))
	for za0002 := range z.Users {
		if z.Users[za0002] == nil {
			o = hsp.AppendNil(o)
		} else {
			// map header, size 2
			o = append(o, 0x82, 0x82)
			if oTemp, err := z.Users[za0002].Address.MarshalHash(); err != nil {
				return nil, err
			} else {
				o = hsp.AppendBytes(o, oTemp)
			}
			o = append(o, 0x82)
			o = hsp.AppendInt32(o, int32(z.Users[za0002].Permission))
		}
	}
	o = append(o, 0x85)
	o = hsp.AppendArrayHeader(o, uint32(len(z.Miners)))
	for za0001 := range z.Miners {
		if oTemp, err := z.Miners[za0001].MarshalHash(); err != nil {
			return nil, err
		} else {
			o = hsp.AppendBytes(o, oTemp)
		}
	}
	o = append(o, 0x85)
	if oTemp, err := z.Owner.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	o = append(o, 0x85)
	if oTemp, err := z.ID.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	o = append(o, 0x85)
	o = hsp.AppendUint64(o, z.Deposit)
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *SQLChainProfile) Msgsize() (s int) {
	s = 1 + 6 + hsp.ArrayHeaderSize
	for za0002 := range z.Users {
		if z.Users[za0002] == nil {
			s += hsp.NilSize
		} else {
			s += 1 + 8 + z.Users[za0002].Address.Msgsize() + 11 + hsp.Int32Size
		}
	}
	s += 7 + hsp.ArrayHeaderSize
	for za0001 := range z.Miners {
		s += z.Miners[za0001].Msgsize()
	}
	s += 6 + z.Owner.Msgsize() + 3 + z.ID.Msgsize() + 8 + hsp.Uint64Size
	return
}

// MarshalHash marshals for hash
func (z SQLChainRole) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	o = hsp.AppendByte(o, byte(z))
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SQLChainRole) Msgsize() (s int) {
	s = hsp.ByteSize
	return
}

// MarshalHash marshals for hash
func (z *SQLChainUser) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 2
	o = append(o, 0x82, 0x82)
	o = hsp.AppendInt32(o, int32(z.Permission))
	o = append(o, 0x82)
	if oTemp, err := z.Address.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *SQLChainUser) Msgsize() (s int) {
	s = 1 + 11 + hsp.Int32Size + 8 + z.Address.Msgsize()
	return
}

// MarshalHash marshals for hash
func (z UserPermission) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	o = hsp.AppendInt32(o, int32(z))
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z UserPermission) Msgsize() (s int) {
	s = hsp.Int32Size
	return
}
