package useraccountservice

// GetUserCountryResponse matches IUserAccountService/GetUserCountry/v1.
type GetUserCountryResponse struct {
	Response UserCountryPayload `json:"response"`
}

// UserCountryPayload is the top-level user-country payload.
type UserCountryPayload struct {
	Country string `json:"country"`
}
