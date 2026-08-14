package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistImageResponse(t *testing.T) {
	basePath := configureLocalFileStorage(t)
	body := []byte(`{
		"created": 1,
		"custom": "keep",
		"data": [{
			"b64_json": "iVBORw0KGgo=",
			"revised_prompt": "prompt",
			"extra": 123
		}]
	}`)

	result, err := PersistImageResponse(context.Background(), body, "")
	require.NoError(t, err)

	var response map[string]json.RawMessage
	require.NoError(t, common.Unmarshal(result, &response))
	assert.JSONEq(t, `"keep"`, string(response["custom"]))

	var items []map[string]json.RawMessage
	require.NoError(t, common.Unmarshal(response["data"], &items))
	require.Len(t, items, 1)
	_, hasBase64 := items[0]["b64_json"]
	assert.False(t, hasBase64)
	assert.JSONEq(t, `123`, string(items[0]["extra"]))

	var fileURL string
	require.NoError(t, common.Unmarshal(items[0]["url"], &fileURL))
	assert.Contains(t, fileURL, "/gc/img/"+time.Now().Format("200601")+"/")
	_, err = os.Stat(publicURLToLocalPath(t, basePath, fileURL))
	require.NoError(t, err)
}

func TestPersistImageResponseKeepsOriginalOnPersistenceFailure(t *testing.T) {
	configureLocalFileStorage(t)
	body := []byte(`{"created":1,"data":[{"b64_json":"%%%","revised_prompt":"keep"}]}`)

	result, err := PersistImageResponse(context.Background(), body, "")
	require.NoError(t, err)
	assert.Equal(t, body, result)
}
