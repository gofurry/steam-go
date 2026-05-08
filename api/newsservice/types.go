package newsservice

// ConvertHTMLToBBCodeResponse matches INewsService/ConvertHTMLToBBCode/v1.
type ConvertHTMLToBBCodeResponse struct {
	Response ConvertHTMLToBBCodePayload `json:"response"`
}

// ConvertHTMLToBBCodePayload is the top-level conversion payload.
type ConvertHTMLToBBCodePayload struct {
	ConvertedContent string `json:"converted_content"`
	FoundHTML        bool   `json:"found_html"`
}
