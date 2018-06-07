// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereumproject/go-ethereum/common"
	"github.com/ethereumproject/go-ethereum/core/state"
	"github.com/ethereumproject/go-ethereum/rpc"
)

func TestPublicEthereumAPI_GasPrice(t *testing.T) {
	type fields struct {
		e      *Ethereum
		gpo    *GasPriceOracle
		Apiish *Apiish
	}
	e := &Ethereum{}
	e.gpo = NewGasPriceOracle(e)
	et := NewPublicEthereumAPI(e)

	tests := []struct {
		name   string
		fields fields
		want   *big.Int
	}{
		// TODO: Add test cases.
		{
			name: "1",
			fields: fields{
				e:      nil,
				gpo:    et.gpo,
				Apiish: et.Apiish,
			},
			want: big.NewInt(0),
		},
		// {
		// 	name: "1",
		// 	fields: {
		// 		e:      nil,
		// 		gpo:    NewGasPriceOracle(nil),
		// 		Apiish: NewPublicGethAPI(nil),
		// 	},
		// 	want: big.NewInt(0),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PublicEthereumAPI{
				e:      tt.fields.e,
				gpo:    tt.fields.gpo,
				Apiish: tt.fields.Apiish,
			}
			if got := s.GasPrice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublicEthereumAPI.GasPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrivateMinerAPI_SetGasPrice(t *testing.T) {
	type fields struct {
		e *Ethereum
	}
	type args struct {
		gasPrice rpc.HexNumber
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PrivateMinerAPI{
				e: tt.fields.e,
			}
			if got := s.SetGasPrice(tt.args.gasPrice); got != tt.want {
				t.Errorf("PrivateMinerAPI.SetGasPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_callmsg_GasPrice(t *testing.T) {
	type fields struct {
		from     *state.StateObject
		to       *common.Address
		gas      *big.Int
		gasPrice *big.Int
		value    *big.Int
		data     []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   *big.Int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := callmsg{
				from:     tt.fields.from,
				to:       tt.fields.to,
				gas:      tt.fields.gas,
				gasPrice: tt.fields.gasPrice,
				value:    tt.fields.value,
				data:     tt.fields.data,
			}
			if got := m.GasPrice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("callmsg.GasPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}
