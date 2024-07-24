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
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

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

func TestHexOrDecimal256(t *testing.T) {
	expected := big.NewInt(31337)
	input := math.NewHexOrDecimal256(31337)
	assert.True(t, (*big.Int)(input).Cmp(expected) == 0)
}

func TestTypedDataArrayValidate(t *testing.T) {

	typedData0 := TypedData{
		Types: Types{
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
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "OrderComponents",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		//Message: TypedDataMessage{
		//	"tree": []interface{}{
		//		[]interface{}{
		//			map[string]interface{}{
		//				"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
		//				"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
		//				"offer": []interface{}{
		//					map[string]interface{}{
		//						"itemType":             "2",
		//						"token":                "0xa62835d1a6bf5f521c4e2746e1f51c923b8f3483",
		//						"identifierOrCriteria": "0",
		//						"startAmount":          "1",
		//						"endAmount":            "1",
		//					},
		//				},
		//				"consideration": []interface{}{
		//					map[string]interface{}{
		//						"itemType":             "0",
		//						"token":                "0x0000000000000000000000000000000000000000",
		//						"identifierOrCriteria": "0",
		//						"startAmount":          "1000000",
		//						"endAmount":            "1000000",
		//						"recipient":            "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
		//					},
		//				},
		//				"orderType":  "2",
		//				"startTime":  "1721370485",
		//				"endTime":    "1784442485",
		//				"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
		//				"salt":       "0xd23957f29d709c45",
		//				"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
		//				"counter":    "0",
		//			},
		//		},
		//	},
		//},
		Message: TypedDataMessage{
			"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
			"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
			"offer": []interface{}{
				map[string]interface{}{
					"itemType":             "2",
					"token":                "0xa62835d1a6bf5f521c4e2746e1f51c923b8f3483",
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
			"startTime":  "1721370485",
			"endTime":    "1784442485",
			"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
			"salt":       "0xd23957f29d709c45",
			"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
			"counter":    "0",
		},
	}

	typedData1 := TypedData{
		Types: Types{
			"BulkOrder": []Type{
				{
					Name: "tree",
					Type: "OrderComponents[2]",
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
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				map[string]interface{}{
					"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
					"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
					"offer": []interface{}{
						map[string]interface{}{
							"itemType":             "2",
							"token":                "0x262e2b50219620226c5fb5956432a88fffd94ba7",
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
					"startTime":  "1721370489",
					"endTime":    "1784442489",
					"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
					"salt":       "0x61bc238c47087001",
					"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
					"counter":    "0",
				},
				map[string]interface{}{
					"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
					"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
					"offer": []interface{}{
						map[string]interface{}{
							"itemType":             "2",
							"token":                "0x262e2b50219620226c5fb5956432a88fffd94ba7",
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
					"startTime":  "1721370489",
					"endTime":    "1784442489",
					"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
					"salt":       "0x9bf89a1fed29e323",
					"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
					"counter":    "0",
				},
			},
		},
	}

	typedData2 := TypedData{
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
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				[]interface{}{
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
						"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
						"offer": []interface{}{
							map[string]interface{}{
								"itemType":             "2",
								"token":                "0x06b3244b086cecc40f1e5a826f736ded68068a0f",
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
						"startTime":  "1721370492",
						"endTime":    "1784442492",
						"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":       "0x143555bae9f6c3dd",
						"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
						"counter":    "0",
					},
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
						"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
						"offer": []interface{}{
							map[string]interface{}{
								"itemType":             "2",
								"token":                "0x06b3244b086cecc40f1e5a826f736ded68068a0f",
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
						"startTime":  "1721370492",
						"endTime":    "1784442492",
						"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":       "0xa40e43309562a29b",
						"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
						"counter":    "0",
					},
				},
				[]interface{}{
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
						"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
						"offer": []interface{}{
							map[string]interface{}{
								"itemType":             "2",
								"token":                "0x06b3244b086cecc40f1e5a826f736ded68068a0f",
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
						"startTime":  "1721370492",
						"endTime":    "1784442492",
						"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
						"salt":       "0x43ef498909096747",
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

	typedData3 := TypedData{
		Types: Types{
			"BulkOrder": []Type{
				{
					Name: "tree",
					Type: "OrderComponents[2][2][2]",
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
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				[]interface{}{
					[]interface{}{
						map[string]interface{}{
							"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
							"offer": []interface{}{
								map[string]interface{}{
									"itemType":             "2",
									"token":                "0x2f8d338360d095a72680a943a22fe6a0d398a0b4",
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
							"startTime":  "1721798594",
							"endTime":    "1784870594",
							"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
							"salt":       "0xea6603a0f70fa487",
							"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
							"counter":    "0",
						},
						map[string]interface{}{
							"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
							"offer": []interface{}{
								map[string]interface{}{
									"itemType":             "2",
									"token":                "0x2f8d338360d095a72680a943a22fe6a0d398a0b4",
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
							"startTime":  "1721798594",
							"endTime":    "1784870594",
							"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
							"salt":       "0x5e5926d7394ebc15",
							"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
							"counter":    "0",
						},
					},
					[]interface{}{
						map[string]interface{}{
							"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
							"offer": []interface{}{
								map[string]interface{}{
									"itemType":             "2",
									"token":                "0x2f8d338360d095a72680a943a22fe6a0d398a0b4",
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
							"startTime":  "1721798594",
							"endTime":    "1784870594",
							"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
							"salt":       "0xecb78d6939c73069",
							"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
							"counter":    "0",
						},
						map[string]interface{}{
							"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
							"offer": []interface{}{
								map[string]interface{}{
									"itemType":             "2",
									"token":                "0x2f8d338360d095a72680a943a22fe6a0d398a0b4",
									"identifierOrCriteria": "3",
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
							"startTime":  "1721798594",
							"endTime":    "1784870594",
							"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
							"salt":       "0xb468f60676944382",
							"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
							"counter":    "0",
						},
					},
				},
				[]interface{}{
					[]interface{}{
						map[string]interface{}{
							"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
							"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
							"offer": []interface{}{
								map[string]interface{}{
									"itemType":             "2",
									"token":                "0x2f8d338360d095a72680a943a22fe6a0d398a0b4",
									"identifierOrCriteria": "4",
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
							"startTime":  "1721798594",
							"endTime":    "1784870594",
							"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
							"salt":       "0x84e6227b4e60e915",
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
					[]interface{}{
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
		},
	}

	typedData4 := TypedData{
		Types: Types{
			"BulkOrder": []Type{
				{
					Name: "tree",
					Type: "OrderComponents[2][2][2][2]",
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
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				[]interface{}{
					[]interface{}{
						[]interface{}{
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x0aea7fa399421d03",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x547bf138eb02083d",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
						},
						[]interface{}{
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x1fdc5cf73c2846",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
										"identifierOrCriteria": "3",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x982c1f45640e7238",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
						},
					},
					[]interface{}{
						[]interface{}{
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
										"identifierOrCriteria": "4",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x52438c99a8a41726",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
										"identifierOrCriteria": "5",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x17a2c2cca1c84c21",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
						},
						[]interface{}{
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
										"identifierOrCriteria": "6",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0xad92568ed1ac612a",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
										"identifierOrCriteria": "7",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x1d72b3b8d1dfd249",
								"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
								"counter":    "0",
							},
						},
					},
				},
				[]interface{}{
					[]interface{}{
						[]interface{}{
							map[string]interface{}{
								"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
								"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
								"offer": []interface{}{
									map[string]interface{}{
										"itemType":             "2",
										"token":                "0xc1eed9232a0a44c2463acb83698c162966fbc78d",
										"identifierOrCriteria": "8",
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
								"startTime":  "1721798609",
								"endTime":    "1784870609",
								"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
								"salt":       "0x03834e51be5d84bf",
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
						[]interface{}{
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
					[]interface{}{
						[]interface{}{
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
						[]interface{}{
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
			},
		},
	}

	typedData5 := TypedData{
		Types: Types{
			"BulkOrder": []Type{
				{
					Name: "tree",
					Type: "OrderComponents[2][2][2][2][2]",
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
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				[]interface{}{
					[]interface{}{
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xad1cd95e33df638d",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x8c46499a985556e7",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xf113dfcd2fd90f34",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "3",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x86e05a39d63439ab",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
						},
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "4",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xea0549310e1fcf98",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "5",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x3c7ebf9ef6b22665",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "6",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x6d8a0266cd32c3cf",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "7",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x5508a5f8c60e761f",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
						},
					},
					[]interface{}{
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "8",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xde53a72c803e4780",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "9",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xb4ad240444cd7c1d",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "10",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xd5559fcc7192a0e8",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "11",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xdde06bfaf059bf21",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
						},
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "12",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x02f0b40bd92ec30d",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "13",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0xa2e93fa3d6add8ef",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "14",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x1c9502d5e53ca013",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "15",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x32c9f5cdbbcf0a0e",
									"conduitKey": "0x0000000000000000000000000000000000000000000000000000000000000000",
									"counter":    "0",
								},
							},
						},
					},
				},
				[]interface{}{
					[]interface{}{
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
									"zone":    "0x84c7fea5b8c328db68632a3bdda3aadab7d36e66",
									"offer": []interface{}{
										map[string]interface{}{
											"itemType":             "2",
											"token":                "0xd28f3246f047efd4059b24fa1fa587ed9fa3e77f",
											"identifierOrCriteria": "16",
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
									"startTime":  "1721798656",
									"endTime":    "1784870656",
									"zoneHash":   "0x0000000000000000000000000000000000000000000000000000000000000000",
									"salt":       "0x9cc16b772ab421d8",
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
							[]interface{}{
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
						[]interface{}{
							[]interface{}{
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
							[]interface{}{
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
					[]interface{}{
						[]interface{}{
							[]interface{}{
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
							[]interface{}{
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
						[]interface{}{
							[]interface{}{
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
							[]interface{}{
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
				},
			},
		},
	}

	typedData2DLean := TypedData{
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
			},
			"EIP712Domain": []Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "BulkOrder",
		Domain: TypedDataDomain{
			Name:              "ImmutableSeaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(31337),
			VerifyingContract: "0x3870289A34bba912a05B2c0503F7484dD18d2f6F",
		},
		Message: TypedDataMessage{
			"tree": []interface{}{
				[]interface{}{
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
					},
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92267",
					},
				},
				[]interface{}{
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92268",
					},
					map[string]interface{}{
						"offerer": "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92269",
					},
				},
			},
		},
	}

	//typedDataEthersExample := TypedData{
	//	Types: Types{
	//		"Struct5": []Type{
	//			{
	//				Name: "param2",
	//				Type: "string[3][3]",
	//			},
	//		},
	//		"EIP712Domain": []Type{
	//			{Name: "name", Type: "string"},
	//			{Name: "chainId", Type: "uint256"},
	//		},
	//	},
	//	PrimaryType: "Struct5",
	//	Domain: TypedDataDomain{
	//		Name:    "Moo é🚀oM o     MMoéMé🚀MéMéM",
	//		ChainId: math.NewHexOrDecimal256(900),
	//	},
	//	Message: TypedDataMessage{
	//		"param2": []interface{}{
	//			[]string{
	//				"Moo é🚀MMo 🚀🚀MM oéoooéé",
	//				"Moo é🚀 éo🚀 🚀oéoMo",
	//				"Moo é🚀oééoéM oMé Moéoo oo M MMoééoooo🚀M🚀o🚀oéMo oo é Moo ooo oo🚀 ",
	//			},
	//			[]string{
	//				"Moo é🚀o🚀ooéMMooooMo oo  o🚀 M🚀Mooo oé🚀o🚀oéoM 🚀M oéo🚀 🚀🚀  🚀o🚀   M",
	//				"Moo é🚀oéé🚀ooéo ooooM🚀🚀éo🚀🚀🚀ooé🚀 éooM🚀oooooMoo Mo🚀ooooMM 🚀 🚀",
	//				"Moo é🚀oo🚀M o🚀oo🚀éoMoooM  oM M🚀ooMM🚀 éo MooMM  éooo",
	//			},
	//			[]string{
	//				"Moo é🚀MMMééo oM o🚀 🚀🚀 Mo o🚀éo🚀oMoé éé oo🚀éé🚀Méoé🚀🚀oéoo 🚀",
	//				"Moo é🚀 🚀M",
	//				"Moo é🚀oo🚀Mo🚀🚀oMo🚀M🚀 o  MMoo   ééMoé MoMoMMooééoo🚀 éo",
	//			},
	//		},
	//	},
	//}

	type fields struct {
		Input TypedData
	}

	type want struct {
		completeHash string
		domainHash   string
		messageHash  string
	}

	tests := map[string]struct {
		Fields fields
		Want   want
	}{
		"obmr-non-array": {
			Fields: fields{
				Input: typedData0,
			}, Want: want{
				completeHash: "0x09311d5cc4e0d26af26c78438f55094fdf489083cd75223073db9a0a5da22b84",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0x81203a474f76afd1376c9f00b3947b5b7e89a73b13b165f999d540377bb1c2fb", // 0x369514af6e781a85186ef2c059e0d9e7b14e5d58a95970655794439eca7f3f7e
			},
		},
		"obmr-single-dimension": {
			Fields: fields{
				Input: typedData1,
			}, Want: want{
				completeHash: "0xa51998e192ae3b3f551e481205b3e84f47041cd1fdccc6ebeb84d09dbaa9163c",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0x449764b1b6c14b5c3d2b69ca5f112e4172feefc49d9712be588862d606d82552", // 0x369514af6e781a85186ef2c059e0d9e7b14e5d58a95970655794439eca7f3f7e
			},
		},
		"obmr-two-dimension": {
			Fields: fields{
				Input: typedData2,
			}, Want: want{
				completeHash: "0x42554635de2cd114d9e36535d7890e93525faa924af52182ae72c069c4909de6",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0x369514af6e781a85186ef2c059e0d9e7b14e5d58a95970655794439eca7f3f7e",
			},
		},
		"obmr-three-dimension": {
			Fields: fields{
				Input: typedData3,
			}, Want: want{
				completeHash: "0xba39e9b22ff1ecd70a6ca86875a8c28e8c3f638d089071edaf97c1e4b6b072a0",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0xb5676629424954a09bd2aea49194741ea05d9525153b062b4507b45c7a0ad759",
			},
		},
		"obmr-four-dimension": {
			Fields: fields{
				Input: typedData4,
			}, Want: want{
				completeHash: "0x9478f43bbeeb2da013ff8595f19a3cb803b41c72596dc0faa6ff1d1b8a70c9f8",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0x6141febc9edb34347a91729533c7c8d3a934fdc34429a0ef9497d5d22664321c",
			},
		},
		"obmr-five-dimension": {
			Fields: fields{
				Input: typedData5,
			}, Want: want{
				completeHash: "0xc319cb3f3ad00993bd33741486db2c9dec09fe8afc1757697cb7ce2b35e1b8e1",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0x127b8c1a34e217116438ea08d10a5e6679ec57fd553ba64891f24500f12dada0",
			},
		},
		"obmr-two-dimension-lean": {
			Fields: fields{
				Input: typedData2DLean,
			}, Want: want{
				completeHash: "0x90c689705d7b249b2c5ff6368a0a3cd8b57aa31b13479c8cf074d85b2416af84",
				domainHash:   "0x94c78e94e233546655365725a17a437f48bb870b898e35b894da4a0887172dc2",
				messageHash:  "0xa931ed014c19242a3e88739106335a65918f8ac748ae4e9965eae8cc2c4c16c7",
			},
		},
		// fails atm because fixed size arrays for primitive types is not supported.
		//"ethers-example": {
		//	Fields: fields{
		//		Input: typedDataEthersExample,
		//	}, Want: want{
		//		completeHash: "0x42bfc8f80f73b02a800f2cf5f3b9b96c6774a43c706758c8f34f1fabf946b001",
		//		domainHash:   "0x5247656f531410c29fada51024987197407dd7082c1280d87ab649e5ab05a646",
		//		messageHash:  "0xa008f077c5a31e71f01d75fbc91d0fdd4c79d37d634a35e99e528d46ad199417",
		//	},
		//},
	}

	for name, tt := range tests {
		tc := tt

		t.Run(name, func(t *testing.T) {
			td := tc.Fields.Input

			fmt.Println("test name: ", name)

			domainSeparator, err := td.HashStruct("EIP712Domain", td.Domain.Map())
			if err != nil {
				t.Fatalf("failed to hash domain separator: %v", err)
			}

			messageHash, err := td.HashStruct(td.PrimaryType, td.Message)
			if err != nil {
				t.Fatalf("failed to hash message: %v", err)
			}

			structured := crypto.Keccak256Hash([]byte(fmt.Sprintf("%s%s%s", "\x19\x01", string(domainSeparator), string(messageHash))))

			assert.Equal(t, tc.Want.completeHash, structured.String(), "structured hashes do not match")
			assert.Equal(t, tc.Want.domainHash, domainSeparator.String(), "domainSeparator hashes do not match")
			assert.Equal(t, tc.Want.messageHash, messageHash.String(), "message hashes do not match")

			if err := td.validate(); err != nil {
				t.Errorf("expected typed data to pass validation, got: %v", err)
			}

			if strings.Contains(name, "obmr") && !strings.Contains(name, "non-array") {
				// Should be able to accept dynamic arrays
				td.Types["BulkOrder"][0].Type = "OrderComponents[]"

				if err := td.validate(); err != nil {
					t.Errorf("expected typed data to pass validation, got: %v", err)
				}

				// Should be able to accept standard types
				td.Types["BulkOrder"][0].Type = "OrderComponents"

				if err := td.validate(); err != nil {
					t.Errorf("expected typed data to pass validation, got: %v", err)
				}

				// Should fail on a bad reference
				td.Types["BulkOrder"][0].Type = "OrderComponent"

				err = td.validate()
				assert.Error(t, err, "expected typed data to fail validation, got nil")
			}
		})

	}
}
