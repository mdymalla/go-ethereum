// Copyright 2019 The go-ethereum Authors
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

package apitypes

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/math"
)

func TestBytesPadding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Type   string
		Input  []byte
		Output []byte // nil => error
	}{
		{
			// Fail on wrong length
			Type:   "bytes20",
			Input:  []byte{},
			Output: nil,
		},
		{
			Type:   "bytes1",
			Input:  []byte{1},
			Output: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			Type:   "bytes1",
			Input:  []byte{1, 2},
			Output: nil,
		},
		{
			Type:   "bytes7",
			Input:  []byte{1, 2, 3, 4, 5, 6, 7},
			Output: []byte{1, 2, 3, 4, 5, 6, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			Type:   "bytes32",
			Input:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			Output: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		},
		{
			Type:   "bytes32",
			Input:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33},
			Output: nil,
		},
	}

	d := TypedData{}
	for i, test := range tests {
		val, err := d.EncodePrimitiveValue(test.Type, test.Input, 1)
		if test.Output == nil {
			if err == nil {
				t.Errorf("test %d: expected error, got no error (result %x)", i, val)
			}
		} else {
			if err != nil {
				t.Errorf("test %d: expected no error, got %v", i, err)
			}
			if len(val) != 32 {
				t.Errorf("test %d: expected len 32, got %d", i, len(val))
			}
			if !bytes.Equal(val, test.Output) {
				t.Errorf("test %d: expected %x, got %x", i, test.Output, val)
			}
		}
	}
}

func TestParseAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		Input  interface{}
		Output []byte // nil => error
	}{
		{
			Input:  [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
			Output: common.FromHex("0x0000000000000000000000000102030405060708090A0B0C0D0E0F1011121314"),
		},
		{
			Input:  "0x0102030405060708090A0B0C0D0E0F1011121314",
			Output: common.FromHex("0x0000000000000000000000000102030405060708090A0B0C0D0E0F1011121314"),
		},
		{
			Input:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
			Output: common.FromHex("0x0000000000000000000000000102030405060708090A0B0C0D0E0F1011121314"),
		},
		// Various error-cases:
		{Input: "0x000102030405060708090A0B0C0D0E0F1011121314"}, // too long string
		{Input: "0x01"}, // too short string
		{Input: ""},
		{Input: [32]byte{}},       // too long fixed-size array
		{Input: [21]byte{}},       // too long fixed-size array
		{Input: make([]byte, 19)}, // too short slice
		{Input: make([]byte, 21)}, // too long slice
		{Input: nil},
	}

	d := TypedData{}
	for i, test := range tests {
		val, err := d.EncodePrimitiveValue("address", test.Input, 1)
		if test.Output == nil {
			if err == nil {
				t.Errorf("test %d: expected error, got no error (result %x)", i, val)
			}
			continue
		}
		if err != nil {
			t.Errorf("test %d: expected no error, got %v", i, err)
		}
		if have, want := len(val), 32; have != want {
			t.Errorf("test %d: have len %d, want %d", i, have, want)
		}
		if !bytes.Equal(val, test.Output) {
			t.Errorf("test %d: want %x, have %x", i, test.Output, val)
		}
	}
}

func TestParseBytes(t *testing.T) {
	t.Parallel()
	for i, tt := range []struct {
		v   interface{}
		exp []byte
	}{
		{"0x", []byte{}},
		{"0x1234", []byte{0x12, 0x34}},
		{[]byte{12, 34}, []byte{12, 34}},
		{hexutil.Bytes([]byte{12, 34}), []byte{12, 34}},
		{"1234", nil},    // not a proper hex-string
		{"0x01233", nil}, // nibbles should be rejected
		{"not a hex string", nil},
		{15, nil},
		{nil, nil},
		{[2]byte{12, 34}, []byte{12, 34}},
		{[8]byte{12, 34, 56, 78, 90, 12, 34, 56}, []byte{12, 34, 56, 78, 90, 12, 34, 56}},
		{[16]byte{12, 34, 56, 78, 90, 12, 34, 56, 12, 34, 56, 78, 90, 12, 34, 56}, []byte{12, 34, 56, 78, 90, 12, 34, 56, 12, 34, 56, 78, 90, 12, 34, 56}},
	} {
		out, ok := parseBytes(tt.v)
		if tt.exp == nil {
			if ok || out != nil {
				t.Errorf("test %d: expected !ok, got ok = %v with out = %x", i, ok, out)
			}
			continue
		}
		if !ok {
			t.Errorf("test %d: expected ok got !ok", i)
		}
		if !bytes.Equal(out, tt.exp) {
			t.Errorf("test %d: expected %x got %x", i, tt.exp, out)
		}
	}
}

func TestParseInteger(t *testing.T) {
	t.Parallel()
	for i, tt := range []struct {
		t   string
		v   interface{}
		exp *big.Int
	}{
		{"uint32", "-123", nil},
		{"int32", "-123", big.NewInt(-123)},
		{"int32", big.NewInt(-124), big.NewInt(-124)},
		{"uint32", "0xff", big.NewInt(0xff)},
		{"int8", "0xffff", nil},
	} {
		res, err := parseInteger(tt.t, tt.v)
		if tt.exp == nil && res == nil {
			continue
		}
		if tt.exp == nil && res != nil {
			t.Errorf("test %d, got %v, expected nil", i, res)
			continue
		}
		if tt.exp != nil && res == nil {
			t.Errorf("test %d, got '%v', expected %v", i, err, tt.exp)
			continue
		}
		if tt.exp.Cmp(res) != 0 {
			t.Errorf("test %d, got %v expected %v", i, res, tt.exp)
		}
	}
}

func TestConvertStringDataToSlice(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b", "c"}
	var it interface{} = slice
	_, err := convertDataToSlice(it)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertUint256DataToSlice(t *testing.T) {
	t.Parallel()
	slice := []*math.HexOrDecimal256{
		math.NewHexOrDecimal256(1),
		math.NewHexOrDecimal256(2),
		math.NewHexOrDecimal256(3),
	}
	var it interface{} = slice
	_, err := convertDataToSlice(it)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertAddressDataToSlice(t *testing.T) {
	t.Parallel()
	slice := []common.Address{
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		common.HexToAddress("0x0000000000000000000000000000000000000002"),
		common.HexToAddress("0x0000000000000000000000000000000000000003"),
	}
	var it interface{} = slice
	_, err := convertDataToSlice(it)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTypedDataArrayValidate(t *testing.T) {
	t.Parallel()

	typedData := TypedData{
		Types: Types{
			"BulkOrder": []Type{
				{
					Name: "tree",
					Type: "OrderComponents[2][2]",
				},
			},
			"OrderComponents": []Type{
				{
					Name: "offerer",
					Type: "address",
				},
				{
					Name: "zone",
					Type: "address",
				},
				{
					Name: "offer",
					Type: "OfferItem[]",
				},
				{
					Name: "consideration",
					Type: "ConsiderationItem[]",
				},
				{
					Name: "orderType",
					Type: "uint8",
				},
				{
					Name: "startTime",
					Type: "uint256",
				},
				{
					Name: "endTime",
					Type: "uint256",
				},
				{
					Name: "zoneHash",
					Type: "bytes32",
				},
				{
					Name: "salt",
					Type: "uint256",
				},
				{
					Name: "conduitKey",
					Type: "bytes32",
				},
				{
					Name: "counter",
					Type: "uint256",
				},
			},
			"OfferItem": []Type{
				{
					Name: "itemType",
					Type: "uint8",
				},
				{
					Name: "token",
					Type: "address",
				},
				{
					Name: "identifierOrCriteria",
					Type: "uint256",
				},
				{
					Name: "startAmount",
					Type: "uint256",
				},
				{
					Name: "endAmount",
					Type: "uint256",
				},
			},
			"ConsiderationItem": []Type{
				{
					Name: "itemType",
					Type: "uint8",
				},
				{
					Name: "token",
					Type: "address",
				},
				{
					Name: "identifierOrCriteria",
					Type: "uint256",
				},
				{
					Name: "startAmount",
					Type: "uint256",
				},
				{
					Name: "endAmount",
					Type: "uint256",
				},
				{
					Name: "recipient",
					Type: "address",
				},
			},
			"EIP712Domain": []Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint32"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x73F43673d3Ca08037e6b6271F801E19277EC1188",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				[]interface{}{
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
						"zone":    "0xf1c8ffe8eedfb5c302b89ccb22b31aedd7ae6043",
						"offer": []interface{}{
							map[string]interface{}{
								"itemType":             "2",
								"token":                "0x9d3999af03458c11c78f7e6c0fae712b455d4e33",
								"identifierOrCriteria": "0",
								"startAmount":          "1",
								"endAmount":            "1",
							},
						},
						"consideration": []interface{}{
							map[string]interface{}{
								"itemType":             "0",
								"token":                "0x0000000000000000000000000000000000000000",
								"identifierOrCriteria": "0",
								"startAmount":          "1000000",
								"endAmount":            "1000000",
								"recipient":            "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							},
						},
						"orderType":  "2",
						"startTime":  "1721344316",
						"endTime":    "1784416316",
						"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":       "0x288351a1835eb797",
						"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
						"counter":    "0",
					},
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
						"zone":    "0xf1c8ffe8eedfb5c302b89ccb22b31aedd7ae6043",
						"offer": []interface{}{
							map[string]interface{}{
								"itemType":             "2",
								"token":                "0x9d3999af03458c11c78f7e6c0fae712b455d4e33",
								"identifierOrCriteria": "1",
								"startAmount":          "1",
								"endAmount":            "1",
							},
						},
						"consideration": []interface{}{
							map[string]interface{}{
								"itemType":             "0",
								"token":                "0x0000000000000000000000000000000000000000",
								"identifierOrCriteria": "0",
								"startAmount":          "1000000",
								"endAmount":            "1000000",
								"recipient":            "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							},
						},
						"orderType":  "2",
						"startTime":  "1721344316",
						"endTime":    "1784416316",
						"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":       "0x44dd80ec356bed1f",
						"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
						"counter":    "0",
					},
				},
				[]interface{}{
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
						"zone":    "0xf1c8ffe8eedfb5c302b89ccb22b31aedd7ae6043",
						"offer": []interface{}{
							map[string]interface{}{
								"itemType":             "2",
								"token":                "0x9d3999af03458c11c78f7e6c0fae712b455d4e33",
								"identifierOrCriteria": "2",
								"startAmount":          "1",
								"endAmount":            "1",
							},
						},
						"consideration": []interface{}{
							map[string]interface{}{
								"itemType":             "0",
								"token":                "0x0000000000000000000000000000000000000000",
								"identifierOrCriteria": "0",
								"startAmount":          "1000000",
								"endAmount":            "1000000",
								"recipient":            "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							},
						},
						"orderType":  "2",
						"startTime":  "1721344316",
						"endTime":    "1784416316",
						"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":       "0x8e9b116f21c29268",
						"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
						"counter":    "0",
					},
					map[string]interface{}{
						"offerer":       "0x0000000000000000000000000000000000000000",
						"zone":          "0x0000000000000000000000000000000000000000",
						"offer":         []interface{}{},
						"consideration": []interface{}{},
						"orderType":     "0",
						"startTime":     "0",
						"endTime":       "0",
						"zoneHash":      "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":          "0",
						"conduitKey":    "0x0000000000000000000000000000000000000000000000000000000000000000",
						"counter":       "0",
					},
				},
			},
		},
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())

	if err != nil {
		t.Fatalf("failed to hash domain separator: %v", err)
	}

	messageHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)

	if err != nil {
		t.Fatalf("failed to hash message: %v", err)
	}

	// this hash is failing
	structredHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("%s%s%s", "\x19\x01", string(domainSeparator), string(messageHash))))
	fmt.Println("structured hash: ", structredHash.Hex())

	if err := typedData.validate(); err != nil {
		t.Errorf("expected typed data to pass validation, got: %v", err)
	}

	// Should be able to accept dynamic arrays
	typedData.Types["BulkOrder"][0].Type = "OrderComponents[]"

	if err := typedData.validate(); err != nil {
		t.Errorf("expected typed data to pass validation, got: %v", err)
	}

	// Should be able to accept standard types
	typedData.Types["BulkOrder"][0].Type = "OrderComponents"

	if err := typedData.validate(); err != nil {
		t.Errorf("expected typed data to pass validation, got: %v", err)
	}

	// Should fail on a bad reference
	typedData.Types["BulkOrder"][0].Type = "OrderComponent"

	if err := typedData.validate(); err == nil {
		t.Errorf("expected typed data to fail validation, got nil")
	}

}
