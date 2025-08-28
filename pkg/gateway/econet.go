package gateway

import (
	"github.com/gofiber/fiber/v2"
)

type EconetGateway struct {
}

type EconetRequest struct {
	Msisdn    string `json:"sourceNumber" xml:"sourceNumber" form:"sourceNumber"`
	ShortCode string `json:"destinationNumber" xml:"destinationNumber" form:"destinationNumber"`
	Message   string `json:"message" xml:"message" form:"name"`
	SessionId string `json:"transactionID" xml:"transactionID" form:"transactionID"`
	Stage     string `json:"stage" xml:"stage" form:"stage"`
	DestinationNumber string `json:"destinationNumber" xml:"destinationNumber" form:"destinationNumber"`
}

type messageResponse struct {
	//XMLName                  xml.Name `xml:"environment"`
	Xmlns                    string `xml:"xmlns,attr"`
	TransactionTime          string `json:"transactionTime" xml:"transactionTime" form:"transactionTime"`
	SessionId                string `json:"transactionID" xml:"transactionID" form:"transactionID"`
	SourceMsisdn             string `json:"sourceNumber" xml:"sourceNumber" form:"sourceNumber"`
	ShortCode                string `json:"destinationNumber" xml:"destinationNumber" form:"destinationNumber"`
	Message                  string `json:"message" xml:"message" form:"message"`
	Stage                    string `json:"stage" xml:"stage" form:"stage"`
	Channel                  string `json:"channel" xml:"channel" form:"channel"`
	ApplicationTransactionID string `json:"applicationTransactionID" xml:"applicationTransactionID" form:"applicationTransactionID"`
	TransactionType          string `json:"transactionType" xml:"transactionType" form:"transactionType"`
}

func NewEconetGateway() Gateway {
	return &EconetGateway{}
}

func (e *EconetGateway) ToRequest(c *fiber.Ctx) (Request, error) {

	er := EconetRequest{}

	err := c.BodyParser(&er)
	if err != nil {
		return Request{}, err
	}

	return Request{
		Message:   er.Message,
		Msisdn:    er.Msisdn,
		SessionId: er.SessionId,
		Stage	:  er.Stage,
		DestinationNumber: er.DestinationNumber,
	}, nil
}

func (e *EconetGateway) ToResponse(r Response) interface{} {
	stage := "session_active"
	if !r.SessionActive {
		stage = "COMPLETE"
	}
	return messageResponse{
		TransactionTime:          "2022-11-05T21:08:44.405Z",
		SessionId:                r.Session,
		SourceMsisdn:             r.Msisdn,
		ShortCode:                r.Msisdn,
		Message:                  r.Message,
		Stage:                    stage,
		Channel:                  "USSD",
		ApplicationTransactionID: r.Session,
		TransactionType:          "MENU_PROCESSING",
		Xmlns:                    "http://econet.co.zw/intergration/messagingSchema",
	}
}

func (e *EconetGateway) Request() Request {
	return Request{}
}

func (e *EconetGateway) Name() string {
	return "econet"
}
