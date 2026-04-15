package response

import (
	"encoding/json"

	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
)

// DecodeJSON parses the response body into the target type.
func DecodeJSON[T any](body []byte) (T, error) {
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return out, sdkerrors.New(sdkerrors.KindDecode, 0, "decode response body failed", body, err)
	}
	return out, nil
}
