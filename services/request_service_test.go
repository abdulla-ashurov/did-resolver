package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cheqd "github.com/cheqd/cheqd-node/x/cheqd/types"
	resource "github.com/cheqd/cheqd-node/x/resource/types"
	"github.com/cheqd/did-resolver/types"
	"github.com/cheqd/did-resolver/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestResolveDIDDoc(t *testing.T) {
	validDIDDoc := utils.ValidDIDDoc()
	validMetadata := utils.ValidMetadata()
	validResource := utils.ValidResource()
	validDIDResolution := types.NewDidDoc(validDIDDoc)
	subtests := []struct {
		name                   string
		ledgerService          utils.MockLedgerService
		resolutionType         types.ContentType
		did                    string
		expectedDID            *types.DidDoc
		expectedMetadata       types.ResolutionDidDocMetadata
		expectedResolutionType types.ContentType
		expectedError          error
	}{
		{
			name:             "successful resolution",
			ledgerService:    utils.NewMockLedgerService(validDIDDoc, validMetadata, validResource),
			resolutionType:   types.DIDJSONLD,
			did:              utils.ValidDid,
			expectedDID:      &validDIDResolution,
			expectedMetadata: types.NewResolutionDidDocMetadata(utils.ValidDid, validMetadata, []*resource.ResourceHeader{validResource.Header}),
			expectedError:    nil,
		},
		{
			name:             "DID not found",
			ledgerService:    utils.NewMockLedgerService(cheqd.Did{}, cheqd.Metadata{}, resource.Resource{}),
			resolutionType:   types.DIDJSONLD,
			did:              utils.ValidDid,
			expectedDID:      nil,
			expectedMetadata: types.ResolutionDidDocMetadata{},
			expectedError:    types.NewNotFoundError(utils.ValidDid, types.DIDJSONLD, nil, false),
		},
	}

	for _, subtest := range subtests {
		t.Run(subtest.name, func(t *testing.T) {
			context, rec := setupContext("/1.0/identifiers/:did", []string{"did"}, []string{subtest.did}, subtest.resolutionType)
			requestService := NewRequestService("cheqd", subtest.ledgerService)

			if (subtest.resolutionType == "" || subtest.resolutionType == types.DIDJSONLD) && subtest.expectedError == nil {
				subtest.expectedDID.Context = []string{types.DIDSchemaJSONLD}
			} else if subtest.expectedDID != nil {
				subtest.expectedDID.Context = nil
			}
			expectedContentType := defineContentType(subtest.expectedResolutionType, subtest.resolutionType)

			err := requestService.ResolveDIDDoc(context)

			if subtest.expectedError != nil {
				require.EqualValues(t, subtest.expectedError.Error(), err.Error())
			} else {
				var resolutionResult types.DidResolution
				unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resolutionResult)
				require.Empty(t, unmarshalErr)
				require.Empty(t, err)
				require.EqualValues(t, subtest.expectedError, err)
				require.EqualValues(t, subtest.expectedDID, resolutionResult.Did)
				require.EqualValues(t, subtest.expectedMetadata, resolutionResult.Metadata)
				require.EqualValues(t, expectedContentType, resolutionResult.ResolutionMetadata.ContentType)
				require.EqualValues(t, expectedContentType, rec.Header().Get("Content-Type"))
			}
		})
	}
}

func defineContentType(expectedContentType types.ContentType, resolutionType types.ContentType) types.ContentType {
	if expectedContentType == "" {
		return resolutionType
	}
	return expectedContentType
}

func setupContext(path string, paramsNames []string, paramsValues []string, resolutionType types.ContentType) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	context.SetPath(path)
	context.SetParamNames(paramsNames...)
	context.SetParamValues(paramsValues...)
	req.Header.Add("accept", string(resolutionType))
	return context, rec
}
