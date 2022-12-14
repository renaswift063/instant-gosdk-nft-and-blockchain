package sdk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CommitMetaData struct {
	CrudType string
	MetaData *ConsolidatedFileMeta
}

type CommitMetaRequest struct {
	CommitMetaData
	status    StatusCallback
	a         *Allocation
	authToken string
	wg        *sync.WaitGroup
}

type CommitMetaResponse struct {
	TxnID    string
	MetaData *ConsolidatedFileMeta
}

func (req *CommitMetaRequest) processCommitMetaRequest() {
	commitMetaDataBytes, err := json.Marshal(req.CommitMetaData)
	if err != nil {
		req.status.CommitMetaCompleted("", "", nil, err)
		return
	}
	commitMetaDataString := string(commitMetaDataBytes)

	nonce := client.GetClient().Nonce
	if nonce != 0 {
		nonce++
	}
	txn := transaction.NewTransactionEntity(client.GetClientID(), blockchain.GetChainID(), client.GetClientPublicKey(), nonce)
	nonce = txn.TransactionNonce
	if nonce < 1 {
		nonce = transaction.Cache.GetNextNonce(txn.ClientID)
	} else {
		transaction.Cache.Set(txn.ClientID, nonce)
	}
	txn.TransactionNonce = nonce

	txn.TransactionData = commitMetaDataString
	txn.TransactionType = transaction.TxnTypeData
	err = txn.ComputeHashAndSign(client.Sign)
	if err != nil {
		req.status.CommitMetaCompleted(commitMetaDataString, "", nil, err)
		return
	}

	transaction.SendTransactionSync(txn, blockchain.GetMiners())
	querySleepTime := time.Duration(blockchain.GetQuerySleepTime()) * time.Second
	sys.Sleep(querySleepTime)
	retries := 0
	var t *transaction.Transaction
	for retries < blockchain.GetMaxTxnQuery() {
		t, err = transaction.VerifyTransaction(txn.Hash, blockchain.GetSharders())
		if err == nil {
			break
		}
		retries++
		sys.Sleep(querySleepTime)
	}

	if err != nil {
		l.Logger.Error("Error verifying the commit transaction", err.Error(), txn.Hash)
		transaction.Cache.Evict(txn.ClientID)
		req.status.CommitMetaCompleted(commitMetaDataString, "", nil, err)
		return
	}
	if t == nil {
		err = errors.New("transaction_validation_failed", "Failed to get the transaction confirmation")
		transaction.Cache.Evict(txn.ClientID)
		req.status.CommitMetaCompleted(commitMetaDataString, "", nil, err)
		return
	}

	if ok := req.updateCommitMetaTxnToBlobbers(t.Hash); ok {
		l.Logger.Info("Updated commitMetaTxnID to all blobbers")
	} else {
		l.Logger.Info("Failed to update commitMetaTxnID to all blobbers")
	}

	commitMetaResponse := &CommitMetaResponse{
		TxnID:    t.Hash,
		MetaData: req.CommitMetaData.MetaData,
	}

	l.Logger.Info("Marshaling commitMetaResponse to bytes")
	commitMetaReponseBytes, err := json.Marshal(commitMetaResponse)
	if err != nil {
		l.Logger.Error("Failed to marshal commitMetaResponse to bytes")
		transaction.Cache.Evict(txn.ClientID)
		req.status.CommitMetaCompleted(commitMetaDataString, "", t, err)
	}

	l.Logger.Info("Converting commitMetaResponse bytes to string")
	commitMetaResponseString := string(commitMetaReponseBytes)

	l.Logger.Info("Commit complete, Calling CommitMetaCompleted callback")
	req.status.CommitMetaCompleted(commitMetaDataString, commitMetaResponseString, t, nil)

	l.Logger.Info("All process done, Calling return")
}

func (req *CommitMetaRequest) updateCommitMetaTxnToBlobbers(txnHash string) bool {
	numList := len(req.a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	rspCh := make(chan bool, numList)
	for i := 0; i < numList; i++ {
		go req.updatCommitMetaTxnToBlobber(req.a.Blobbers[i], i, txnHash, rspCh)
	}
	req.wg.Wait()
	count := 0
	for i := 0; i < numList; i++ {
		resp := <-rspCh
		if resp {
			count++
		}
	}
	return count == numList
}

func (req *CommitMetaRequest) updatCommitMetaTxnToBlobber(blobber *blockchain.StorageNode, blobberIdx int, txnHash string, rspCh chan<- bool) {

	defer req.wg.Done()
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	formWriter.WriteField("path_hash", req.MetaData.LookupHash)
	formWriter.WriteField("txn_id", txnHash)

	if len(req.authToken) > 0 {
		sEnc, err := base64.StdEncoding.DecodeString(req.authToken)
		if err != nil {
			l.Logger.Error("auth_ticket_decode_error", "Error decoding the auth ticket."+err.Error())
			return
		}
		formWriter.WriteField("auth_token", string(sEnc))
	}

	formWriter.Close()
	httpreq, err := zboxutil.NewCommitMetaTxnRequest(blobber.Baseurl, req.a.Tx, body)
	if err != nil {
		l.Logger.Error("Update commit meta txn request error: ", err.Error())
		return
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.a.ctx, (time.Second * 30))

	zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Update CommitMetaTxn : ", err)
			rspCh <- false
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			rspCh <- true
		} else {
			rspCh <- false
		}
		return err
	})
}
