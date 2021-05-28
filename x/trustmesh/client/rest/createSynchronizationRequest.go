package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	baseledgerTypes "github.com/unibrightio/baseledger/x/baseledger/types"
	"github.com/unibrightio/baseledger/x/trustmesh/proxy"
	"github.com/unibrightio/baseledger/x/trustmesh/types"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type createSynchronizationRequest struct {
	BaseReq                              rest.BaseReq `json:"base_req"`
	Creator                              string       `json:"creator"`
	CreatorName                          string       `json:"creatorName"`
	WorkgroupId                          string       `json:"workgroup_id"`
	Recipient                            string       `json:"recipient"`
	WorkstepType                         string       `json:"workstep_type"`
	BusinessObjectType                   string       `json:"business_object_type"`
	BaseledgerBusinessObjectId           string       `json:"baseledger_business_object_id"`
	BusinessObject                       string       `json:"business_object"`
	ReferencedBaseledgerBusinessObjectId string       `json:"referenced_baseledger_business_object_id"`
}

func createSynchronizationRequestHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createSynchronizationRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddress, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		createSyncReq := &types.SynchronizationRequest{
			WorkgroupId:                          req.WorkgroupId,
			Recipient:                            req.Recipient,
			WorkstepType:                         req.WorkstepType,
			BusinessObjectType:                   req.BusinessObjectType,
			BaseledgerBusinessObjectId:           req.BaseledgerBusinessObjectId,
			BusinessObject:                       req.BusinessObject,
			ReferencedBaseledgerBusinessObjectId: req.ReferencedBaseledgerBusinessObjectId,
		}

		payload, transactionId := proxy.SynchronizeBusinessObject(createSyncReq)

		msg := baseledgerTypes.NewMsgCreateBaseledgerTransaction(clientCtx.GetFromAddress().String(), transactionId, string(payload))
		msg.Creator = baseReq.From
		if err := msg.ValidateBasic(); err != nil {
			fmt.Printf("msg validate basic failed %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		fmt.Printf("msg with encrypted payload to be broadcasted %s\n", msg)

		accNum, accSeq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, fromAddress)

		if err != nil {
			fmt.Printf("error while retrieving acc %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		fmt.Printf("retrieved account %v %v\n", accNum, accSeq)

		kr, err := keyring.New("baseledger", "test", "~/.baseledger", nil)

		if err != nil {
			fmt.Printf("error fetching test key ring %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		key, err := kr.Key(req.CreatorName)

		if err != nil {
			fmt.Printf("error when getting key by name %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		fmt.Printf("key found for address %v\n", key.GetAddress().String())

		clientCtx = clientCtx.
			WithKeyring(kr).
			WithFromAddress(fromAddress).
			WithSkipConfirmation(true).
			WithFromName(req.CreatorName)
			// TODO: check this with team
			// WithBroadcastMode("sync")

		fmt.Printf("broadcast mode %v\n", clientCtx.BroadcastMode)

		txFactory := tx.Factory{}.
			WithChainID(clientCtx.ChainID).
			WithGas(100000).
			WithTxConfig(clientCtx.TxConfig).
			WithAccountNumber(accNum).
			WithSequence(accSeq).
			WithAccountRetriever(clientCtx.AccountRetriever).
			WithKeybase(clientCtx.Keyring)

		txFactory, err = tx.PrepareFactory(clientCtx, txFactory)
		if err != nil {
			fmt.Printf("prepare factory error %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		transaction, err := tx.BuildUnsignedTx(txFactory, msg)
		if err != nil {
			fmt.Printf("build unsigned tx error %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		err = tx.Sign(txFactory, clientCtx.GetFromName(), transaction, true)
		if err != nil {
			fmt.Printf("sign tx error %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		txBytes, err := clientCtx.TxConfig.TxEncoder()(transaction.GetTx())
		if err != nil {
			fmt.Printf("tx encoder %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		// broadcast to a Tendermint node
		res, err := clientCtx.BroadcastTx(txBytes)
		if err != nil {
			fmt.Printf("error while broadcasting tx %v\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		fmt.Printf("TRANSACTION BROADCASTED WITH RESULT %v %v\n", res, res.Timestamp)

		// TODO: fix this mocked entry with proper entries, discuss unkown props with team
		trustmeshEntry := &types.TrustmeshEntry{
			// TODO: what should we use here? this is only availabe in "block" broadcast mode which is not recommended because of timeout
			TendermintBlockId: strconv.FormatUint(uint64(res.Height), 10),
			// TODO: what should we use here?
			TendermintTransactionId: res.TxHash,
			// TODO: what should we use here? timestamp not available
			TendermintTransactionTimestamp:       "2021-05-28T21:42:59.1948424Z",
			Sender:                               "123",
			Receiver:                             "123",
			WorkgroupId:                          req.WorkgroupId,
			WorkstepType:                         req.WorkstepType,
			BaseledgerTransactionType:            "Suggest",
			BaseledgerTransactionId:              transactionId,
			ReferencedBaseledgerTransactionId:    "123",
			BusinessObjectType:                   req.BusinessObjectType,
			BaseledgerBusinessObjectId:           req.BaseledgerBusinessObjectId,
			ReferencedBaseledgerBusinessObjectId: req.ReferencedBaseledgerBusinessObjectId,
			OffchainProcessMessageId:             "123",
			ReferencedProcessMessageId:           "123",
		}

		if !trustmeshEntry.Create() {
			fmt.Printf("error when creating new trustmesh entry")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
