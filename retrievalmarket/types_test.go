package retrievalmarket_test

import (
	"bytes"
	"testing"

	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/shared"
	tut "github.com/filecoin-project/go-fil-markets/shared_testutil"
)

func TestParamsMarshalUnmarshal(t *testing.T) {
	pieceCid := tut.GenerateCids(1)[0]

	allSelector := shared.AllSelector()
	params, err := retrievalmarket.NewParamsV1(abi.NewTokenAmount(123), 456, 789, allSelector, &pieceCid, big.Zero())
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	err = params.MarshalCBOR(buf)
	assert.NoError(t, err)

	unmarshalled := &retrievalmarket.Params{}
	err = unmarshalled.UnmarshalCBOR(buf)
	assert.NoError(t, err)

	assert.Equal(t, params, *unmarshalled)

	nb := basicnode.Prototype.Any.NewBuilder()
	err = dagcbor.Decoder(nb, bytes.NewBuffer(unmarshalled.Selector.Raw))
	assert.NoError(t, err)
	sel := nb.Build()
	assert.Equal(t, sel, allSelector)
}

func TestParamsIntervalBounds(t *testing.T) {
	testCases := []struct {
		name             string
		currentInterval  uint64
		paymentInterval  uint64
		intervalIncrease uint64
		expLowerBound    uint64
		expNextInterval  uint64
	}{{
		currentInterval:  0,
		paymentInterval:  10,
		intervalIncrease: 5,
		expLowerBound:    0,
		expNextInterval:  10,
	}, {
		currentInterval:  10,
		paymentInterval:  10,
		intervalIncrease: 5,
		expLowerBound:    0,
		expNextInterval:  25, // 10 + (10 + 5)
	}, {
		currentInterval:  25,
		paymentInterval:  10,
		intervalIncrease: 5,
		expLowerBound:    10,
		expNextInterval:  45, // 10 + (10 + 5) + (10 + 5 + 5)
	}, {
		currentInterval:  45,
		paymentInterval:  10,
		intervalIncrease: 5,
		expLowerBound:    25,
		expNextInterval:  70, // 10 + (10 + 5) + (10 + 5 + 5) + (10 + 5 + 5 + 5)
	}}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			params := retrievalmarket.Params{
				PaymentInterval:         tc.paymentInterval,
				PaymentIntervalIncrease: tc.intervalIncrease,
			}
			lowerBound := params.IntervalLowerBound(tc.currentInterval)
			nextInterval := params.NextInterval(tc.currentInterval)

			require.Equal(t, tc.expLowerBound, lowerBound)
			require.Equal(t, tc.expNextInterval, nextInterval)
		})
	}
}
