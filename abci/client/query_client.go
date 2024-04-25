package abcicli

import (
	types "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/service"
	cmtsync "github.com/tendermint/tendermint/libs/sync"
)

var _ Client = (*queryClient)(nil)

// queryClient is a variation of localClient which allows queries to run
// concurrently with other operations.
//
// Most state-modifying operations take an exclusive lock on the mtx, as before.
//
// Query-ish operations (Query, CheckTx) take a read-init lock on the queryMtx.
//
// A few state-modifying operations  must also block queries too, so they will
// take the lock on mtx, then a write-lock on queryMtx. These operations are
// SetOption
// InitChain
// Commit - sets the Initialized state
// ApplySnapshotChunk
//
// When grabbing both mutexes, mtx must be taken first, then queryMtx.
// If only grabbing one at a time, either may be taken.
//
// For safe initialization of the queryMtx, its read lock must be held to call
// IsInitialized(), and its write lock must be held when calling Initialize().
//
// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
//
// NOTE: Keep in sync with changes to local_client.
type queryClient struct {
	service.BaseService

	// This mutext must be held for state-modifying calls.
	mtx *cmtsync.Mutex

	// Obtain a read-init lock when calling Application methods that result in a
	// state read.  This is currently:
	// CheckTx
	// DeliverTx
	// Query
	//
	// Only obtain a write lock when calling Application methods that are expected
	// to result in a state mutation.  This is currently:
	// SetOption
	// InitChain
	// Commit - sets the Initialized state
	// ApplySnapshotChunk
	queryMtx *cmtsync.RWInitMutex

	types.Application
	Callback
}

func NewQueryClient(mtx *cmtsync.Mutex, queryMtx *cmtsync.RWInitMutex, app types.Application) Client {
	if mtx == nil {
		mtx = new(cmtsync.Mutex)
	}
	if queryMtx == nil {
		queryMtx = cmtsync.NewRWInitMutex()
	}
	cli := &queryClient{
		mtx:         mtx,
		queryMtx:    queryMtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "committingClient", cli)
	return cli
}

func (app *queryClient) SetResponseCallback(cb Callback) {
	// Need to block all readers
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

// TODO: change types.Application to include Error()?
func (app *queryClient) Error() error {
	return nil
}

func (app *queryClient) FlushAsync() *ReqRes {
	// Do nothing
	return newLocalReqRes(types.ToRequestFlush(), nil)
}

func (app *queryClient) EchoAsync(msg string) *ReqRes {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	return app.callback(
		types.ToRequestEcho(msg),
		types.ToResponseEcho(msg),
	)
}

func (app *queryClient) InfoAsync(req types.RequestInfo) *ReqRes {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.Info(req)
	return app.callback(
		types.ToRequestInfo(req),
		types.ToResponseInfo(res),
	)
}

func (app *queryClient) SetOptionAsync(req types.RequestSetOption) *ReqRes {
	// Need to block all readers
	app.mtx.Lock()
	defer app.mtx.Unlock()
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.SetOption(req)
	return app.callback(
		types.ToRequestSetOption(req),
		types.ToResponseSetOption(res),
	)
}

func (app *queryClient) DeliverTxAsync(params types.RequestDeliverTx) *ReqRes {
	// wait for initialization, then release
	func() {
		app.queryMtx.RInitLock()
		defer app.queryMtx.RInitUnlock()
	}()
	// Lock vs other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(params)
	return app.callback(
		types.ToRequestDeliverTx(params),
		types.ToResponseDeliverTx(res),
	)
}

func (app *queryClient) CheckTxAsync(req types.RequestCheckTx) *ReqRes {
	// Blocked until state is initialized, then by state writers
	app.queryMtx.RInitLock()
	defer app.queryMtx.RInitUnlock()

	res := app.Application.CheckTx(req)
	return app.callback(
		types.ToRequestCheckTx(req),
		types.ToResponseCheckTx(res),
	)
}

func (app *queryClient) QueryAsync(req types.RequestQuery) *ReqRes {
	// Blocked until state is initialized, then by state writers
	app.queryMtx.RInitLock()
	defer app.queryMtx.RInitUnlock()

	res := app.Application.Query(req)
	return app.callback(
		types.ToRequestQuery(req),
		types.ToResponseQuery(res),
	)
}

func (app *queryClient) CommitAsync() *ReqRes {
	// block other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.Commit()
	app.queryMtx.Initialize()

	return app.callback(
		types.ToRequestCommit(),
		types.ToResponseCommit(res),
	)
}

func (app *queryClient) InitChainAsync(req types.RequestInitChain) *ReqRes {
	// block other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.InitChain(req)
	return app.callback(
		types.ToRequestInitChain(req),
		types.ToResponseInitChain(res),
	)
}

func (app *queryClient) BeginBlockAsync(req types.RequestBeginBlock) *ReqRes {
	if len(req.Header.AppHash) > 0 && !app.queryMtx.IsInitialized() {
		app.queryMtx.RLock()
		initialized := app.queryMtx.IsInitialized()
		app.queryMtx.RUnlock()

		if !initialized {
			// We already have some state, so mark as initialized.
			app.queryMtx.Lock()
			app.queryMtx.Initialize()
			app.queryMtx.Unlock()
		}
	}

	// block other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return app.callback(
		types.ToRequestBeginBlock(req),
		types.ToResponseBeginBlock(res),
	)
}

func (app *queryClient) EndBlockAsync(req types.RequestEndBlock) *ReqRes {
	// block other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return app.callback(
		types.ToRequestEndBlock(req),
		types.ToResponseEndBlock(res),
	)
}

func (app *queryClient) ListSnapshotsAsync(req types.RequestListSnapshots) *ReqRes {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.ListSnapshots(req)
	return app.callback(
		types.ToRequestListSnapshots(req),
		types.ToResponseListSnapshots(res),
	)
}

func (app *queryClient) OfferSnapshotAsync(req types.RequestOfferSnapshot) *ReqRes {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.OfferSnapshot(req)
	return app.callback(
		types.ToRequestOfferSnapshot(req),
		types.ToResponseOfferSnapshot(res),
	)
}

func (app *queryClient) LoadSnapshotChunkAsync(req types.RequestLoadSnapshotChunk) *ReqRes {
	// block other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()
	// XXX grab queryMtx write lock to prevent motion blur on readers?

	res := app.Application.LoadSnapshotChunk(req)
	return app.callback(
		types.ToRequestLoadSnapshotChunk(req),
		types.ToResponseLoadSnapshotChunk(res),
	)
}

func (app *queryClient) ApplySnapshotChunkAsync(req types.RequestApplySnapshotChunk) *ReqRes {
	// block other writers
	app.mtx.Lock()
	defer app.mtx.Unlock()

	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.ApplySnapshotChunk(req)
	return app.callback(
		types.ToRequestApplySnapshotChunk(req),
		types.ToResponseApplySnapshotChunk(res),
	)
}

// XXX TODO make corresponding changes to sync versions.

//-------------------------------------------------------

func (app *queryClient) FlushSync() error {
	// Never blocked
	return nil
}

func (app *queryClient) EchoSync(msg string) (*types.ResponseEcho, error) {
	// Never blocked
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *queryClient) InfoSync(req types.RequestInfo) (*types.ResponseInfo, error) {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.Info(req)
	return &res, nil
}

func (app *queryClient) SetOptionSync(req types.RequestSetOption) (*types.ResponseSetOption, error) {
	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.SetOption(req)
	return &res, nil
}

func (app *queryClient) DeliverTxSync(req types.RequestDeliverTx) (*types.ResponseDeliverTx, error) {
	// Blocked until state is initialized, then by state writers
	app.queryMtx.RInitLock()
	defer app.queryMtx.RInitUnlock()

	res := app.Application.DeliverTx(req)
	return &res, nil
}

func (app *queryClient) CheckTxSync(req types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	// Blocked until state is initialized, then by state writers
	app.queryMtx.RInitLock()
	defer app.queryMtx.RInitUnlock()

	res := app.Application.CheckTx(req)
	return &res, nil
}

func (app *queryClient) QuerySync(req types.RequestQuery) (*types.ResponseQuery, error) {
	// Blocked until state is initialized, then by state writers
	app.queryMtx.RInitLock()
	defer app.queryMtx.RInitUnlock()

	res := app.Application.Query(req)
	return &res, nil
}

func (app *queryClient) CommitSync() (*types.ResponseCommit, error) {
	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.Commit()
	app.queryMtx.Initialize()

	return &res, nil
}

func (app *queryClient) InitChainSync(req types.RequestInitChain) (*types.ResponseInitChain, error) {
	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.InitChain(req)
	return &res, nil
}

func (app *queryClient) BeginBlockSync(req types.RequestBeginBlock) (*types.ResponseBeginBlock, error) {
	if len(req.Header.AppHash) > 0 && !app.queryMtx.IsInitialized() {
		// We already have some state, so mark as initialized.
		app.queryMtx.Lock()
		app.queryMtx.Initialize()
		app.queryMtx.Unlock()
	}

	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.BeginBlock(req)
	return &res, nil
}

func (app *queryClient) EndBlockSync(req types.RequestEndBlock) (*types.ResponseEndBlock, error) {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.EndBlock(req)
	return &res, nil
}

func (app *queryClient) ListSnapshotsSync(req types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.ListSnapshots(req)
	return &res, nil
}

func (app *queryClient) OfferSnapshotSync(req types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error) {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.OfferSnapshot(req)
	return &res, nil
}

func (app *queryClient) LoadSnapshotChunkSync(
	req types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error) {
	// Blocked only by state writers
	app.queryMtx.RLock()
	defer app.queryMtx.RUnlock()

	res := app.Application.LoadSnapshotChunk(req)
	return &res, nil
}

func (app *queryClient) ApplySnapshotChunkSync(
	req types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error) {
	// Need to block all readers
	app.queryMtx.Lock()
	defer app.queryMtx.Unlock()

	res := app.Application.ApplySnapshotChunk(req)
	return &res, nil
}

//-------------------------------------------------------

func (app *queryClient) callback(req *types.Request, res *types.Response) *ReqRes {
	// Never blocked
	app.Callback(req, res)
	rr := newLocalReqRes(req, res)
	rr.callbackInvoked = true
	return rr
}
