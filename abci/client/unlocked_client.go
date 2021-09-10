package abcicli

import (
	types "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/service"
)

var _ Client = (*unlockedClient)(nil)

type unlockedClient struct {
	service.BaseService

	types.Application
	Callback
}

func NewUnlockedClient(app types.Application) Client {
	cli := &unlockedClient{
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "unlockedClient", cli)
	return cli
}

func (app *unlockedClient) SetResponseCallback(cb Callback) {
	app.Callback = cb
}

// TODO: change types.Application to include Error()?
func (app *unlockedClient) Error() error {
	return nil
}

func (app *unlockedClient) FlushAsync() *ReqRes {
	// Do nothing
	return newLocalReqRes(types.ToRequestFlush(), nil)
}

func (app *unlockedClient) EchoAsync(msg string) *ReqRes {
	return app.callback(
		types.ToRequestEcho(msg),
		types.ToResponseEcho(msg),
	)
}

func (app *unlockedClient) InfoAsync(req types.RequestInfo) *ReqRes {
	res := app.Application.Info(req)
	return app.callback(
		types.ToRequestInfo(req),
		types.ToResponseInfo(res),
	)
}

func (app *unlockedClient) SetOptionAsync(req types.RequestSetOption) *ReqRes {
	res := app.Application.SetOption(req)
	return app.callback(
		types.ToRequestSetOption(req),
		types.ToResponseSetOption(res),
	)
}

func (app *unlockedClient) DeliverTxAsync(params types.RequestDeliverTx) *ReqRes {
	res := app.Application.DeliverTx(params)
	return app.callback(
		types.ToRequestDeliverTx(params),
		types.ToResponseDeliverTx(res),
	)
}

func (app *unlockedClient) CheckTxAsync(req types.RequestCheckTx) *ReqRes {
	res := app.Application.CheckTx(req)
	return app.callback(
		types.ToRequestCheckTx(req),
		types.ToResponseCheckTx(res),
	)
}

func (app *unlockedClient) QueryAsync(req types.RequestQuery) *ReqRes {
	res := app.Application.Query(req)
	return app.callback(
		types.ToRequestQuery(req),
		types.ToResponseQuery(res),
	)
}

func (app *unlockedClient) CommitAsync() *ReqRes {
	res := app.Application.Commit()
	return app.callback(
		types.ToRequestCommit(),
		types.ToResponseCommit(res),
	)
}

func (app *unlockedClient) InitChainAsync(req types.RequestInitChain) *ReqRes {
	res := app.Application.InitChain(req)
	return app.callback(
		types.ToRequestInitChain(req),
		types.ToResponseInitChain(res),
	)
}

func (app *unlockedClient) BeginBlockAsync(req types.RequestBeginBlock) *ReqRes {
	res := app.Application.BeginBlock(req)
	return app.callback(
		types.ToRequestBeginBlock(req),
		types.ToResponseBeginBlock(res),
	)
}

func (app *unlockedClient) EndBlockAsync(req types.RequestEndBlock) *ReqRes {
	res := app.Application.EndBlock(req)
	return app.callback(
		types.ToRequestEndBlock(req),
		types.ToResponseEndBlock(res),
	)
}

func (app *unlockedClient) ListSnapshotsAsync(req types.RequestListSnapshots) *ReqRes {
	res := app.Application.ListSnapshots(req)
	return app.callback(
		types.ToRequestListSnapshots(req),
		types.ToResponseListSnapshots(res),
	)
}

func (app *unlockedClient) OfferSnapshotAsync(req types.RequestOfferSnapshot) *ReqRes {
	res := app.Application.OfferSnapshot(req)
	return app.callback(
		types.ToRequestOfferSnapshot(req),
		types.ToResponseOfferSnapshot(res),
	)
}

func (app *unlockedClient) LoadSnapshotChunkAsync(req types.RequestLoadSnapshotChunk) *ReqRes {
	res := app.Application.LoadSnapshotChunk(req)
	return app.callback(
		types.ToRequestLoadSnapshotChunk(req),
		types.ToResponseLoadSnapshotChunk(res),
	)
}

func (app *unlockedClient) ApplySnapshotChunkAsync(req types.RequestApplySnapshotChunk) *ReqRes {
	res := app.Application.ApplySnapshotChunk(req)
	return app.callback(
		types.ToRequestApplySnapshotChunk(req),
		types.ToResponseApplySnapshotChunk(res),
	)
}

//-------------------------------------------------------

func (app *unlockedClient) FlushSync() error {
	return nil
}

func (app *unlockedClient) EchoSync(msg string) (*types.ResponseEcho, error) {
	return &types.ResponseEcho{Message: msg}, nil
}

func (app *unlockedClient) InfoSync(req types.RequestInfo) (*types.ResponseInfo, error) {
	res := app.Application.Info(req)
	return &res, nil
}

func (app *unlockedClient) SetOptionSync(req types.RequestSetOption) (*types.ResponseSetOption, error) {
	res := app.Application.SetOption(req)
	return &res, nil
}

func (app *unlockedClient) DeliverTxSync(req types.RequestDeliverTx) (*types.ResponseDeliverTx, error) {
	res := app.Application.DeliverTx(req)
	return &res, nil
}

func (app *unlockedClient) CheckTxSync(req types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	res := app.Application.CheckTx(req)
	return &res, nil
}

func (app *unlockedClient) QuerySync(req types.RequestQuery) (*types.ResponseQuery, error) {
	res := app.Application.Query(req)
	return &res, nil
}

func (app *unlockedClient) CommitSync() (*types.ResponseCommit, error) {
	res := app.Application.Commit()
	return &res, nil
}

func (app *unlockedClient) InitChainSync(req types.RequestInitChain) (*types.ResponseInitChain, error) {
	res := app.Application.InitChain(req)
	return &res, nil
}

func (app *unlockedClient) BeginBlockSync(req types.RequestBeginBlock) (*types.ResponseBeginBlock, error) {
	res := app.Application.BeginBlock(req)
	return &res, nil
}

func (app *unlockedClient) EndBlockSync(req types.RequestEndBlock) (*types.ResponseEndBlock, error) {
	res := app.Application.EndBlock(req)
	return &res, nil
}

func (app *unlockedClient) ListSnapshotsSync(req types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
	res := app.Application.ListSnapshots(req)
	return &res, nil
}

func (app *unlockedClient) OfferSnapshotSync(req types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error) {
	res := app.Application.OfferSnapshot(req)
	return &res, nil
}

func (app *unlockedClient) LoadSnapshotChunkSync(
	req types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error) {
	res := app.Application.LoadSnapshotChunk(req)
	return &res, nil
}

func (app *unlockedClient) ApplySnapshotChunkSync(
	req types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error) {
	res := app.Application.ApplySnapshotChunk(req)
	return &res, nil
}

//-------------------------------------------------------

func (app *unlockedClient) callback(req *types.Request, res *types.Response) *ReqRes {
	app.Callback(req, res)
	return newLocalReqRes(req, res)
}
