package web

import (
	"net/http"

	"tesla-ble-proxy/pkg/tesla"

	"github.com/gin-gonic/gin"
)

type CarCredentialsRequest struct {
	PrivateKey []byte `json:"privateKey"`
	VIN        string `json:"vin"`
}

type SetChargingAmpsRequest struct {
	Amps int32 `json:"amps"`
}

type ChangeChargeLimitRequest struct {
	ChargeLimitPercent int32 `json:"chargeLimitPercent"`
}

type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
}

type ResponseCode string

const ResponseCredentialsNeeded ResponseCode = "credentials-needed"
const ResponseError ResponseCode = "error"
const ResponseOK ResponseCode = "success"

func Serve() {
	r := gin.Default()

	r.POST("/car-credentials", func(c *gin.Context) {
		var creds CarCredentialsRequest

		if err := c.BindJSON(&creds); err != nil {
			c.JSON(http.StatusInternalServerError, &Response{
				Code:    ResponseError,
				Message: err.Error(),
			})

			return
		}

		conn, err := tesla.MakeConnection([]byte(creds.PrivateKey), creds.VIN)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &Response{
				Code:    ResponseError,
				Message: err.Error(),
			})

			return
		}

		connection = conn

		c.Status(http.StatusOK)
	})

	r.POST("/wakeup", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		return conn.Wakeup()
	}))

	r.POST("/lock", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		return conn.Lock()
	}))

	r.POST("/unlock", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		return conn.Unlock()
	}))

	r.POST("/charge-start", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		return conn.ChargeStart()
	}))

	r.POST("/charge-stop", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		return conn.ChargeStop()
	}))

	r.POST("/set-charging-amps", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		var req SetChargingAmpsRequest
		if err := c.BindJSON(&req); err != nil {
			return err
		}

		return conn.SetChargingAmps(req.Amps)
	}))

	r.POST("/change-charge-limit", makeCommandHandler(func(c *gin.Context, conn *tesla.Connection) error {
		var req ChangeChargeLimitRequest
		if err := c.BindJSON(&req); err != nil {
			return err
		}
		return conn.ChangeChargeLimit(req.ChargeLimitPercent)
	}))

	r.Run()
}

func makeCommandHandler(commandHandler func(c *gin.Context, conn *tesla.Connection) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		conn := connection
		if conn == nil {
			c.JSON(http.StatusUnauthorized, &Response{
				Code:    ResponseCredentialsNeeded,
				Message: "You need to provide car credentials by calling /car-credentials API",
			})

			return
		}

		if err := commandHandler(c, conn); err != nil {
			c.JSON(http.StatusInternalServerError, &Response{
				Code:    ResponseError,
				Message: err.Error(),
			})

			return
		}

		c.Status(http.StatusOK)
	}
}

var connection *tesla.Connection
