package web

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teslamotors/vehicle-command/pkg/cache"
	"github.com/teslamotors/vehicle-command/pkg/connector/ble"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
	"github.com/teslamotors/vehicle-command/pkg/vehicle"
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

		connection = MakeConnection([]byte(creds.PrivateKey), creds.VIN)

		c.Status(http.StatusOK)
	})

	r.POST("/wakeup", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		return conn.Wakeup()
	}))

	r.POST("/lock", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		return conn.Lock()
	}))

	r.POST("/unlock", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		return conn.Unlock()
	}))

	r.POST("/charge-start", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		return conn.ChargeStart()
	}))

	r.POST("/charge-stop", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		return conn.ChargeStop()
	}))

	r.POST("/set-charging-amps", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		var req SetChargingAmpsRequest
		if err := c.BindJSON(&req); err != nil {
			return err
		}

		return conn.SetChargingAmps(req.Amps)
	}))

	r.POST("/change-charge-limit", makeCommandHandler(func(c *gin.Context, conn *Connection) error {
		var req ChangeChargeLimitRequest
		if err := c.BindJSON(&req); err != nil {
			return err
		}
		return conn.ChangeChargeLimit(req.ChargeLimitPercent)
	}))

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func makeCommandHandler(commandHandler func(c *gin.Context, conn *Connection) error) func(c *gin.Context) {
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

var connection *Connection

const commandTimeout = 30 * time.Second

type Connection struct {
	privKey      protocol.ECDHPrivateKey
	vin          string
	sessionCache *cache.SessionCache
	commandLock  sync.Mutex
}

func MakeConnection(privateKey []byte, vin string) *Connection {
	return &Connection{
		privKey:      protocol.UnmarshalECDHPrivateKey(privateKey),
		vin:          vin,
		sessionCache: cache.New(0),
	}
}

func (conn *Connection) runCommand(callback func(ctx context.Context, car *vehicle.Vehicle) error) error {
	conn.commandLock.Lock()
	defer conn.commandLock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	connector, err := ble.NewConnection(ctx, conn.vin)
	if err != nil {
		return err
	}
	defer connector.Close()

	car, err := vehicle.NewVehicle(connector, conn.privKey, conn.sessionCache)
	if err != nil {
		return err
	}

	if err := car.Connect(ctx); err != nil {
		return err
	}
	defer car.Disconnect()

	if err := car.StartSession(ctx, nil); err != nil {
		return err
	}
	car.UpdateCachedSessions(conn.sessionCache)

	return callback(ctx, car)
}

func (conn *Connection) Wakeup() error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.Wakeup(ctx)
	})
}

func (conn *Connection) Lock() error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.Lock(ctx)
	})
}

func (conn *Connection) Unlock() error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.Unlock(ctx)
	})
}

func (conn *Connection) ChargeStart() error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.ChargeStart(ctx)
	})
}

func (conn *Connection) ChargeStop() error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.ChargeStop(ctx)
	})
}

func (conn *Connection) SetChargingAmps(amps int32) error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.SetChargingAmps(ctx, amps)
	})
}

func (conn *Connection) ChangeChargeLimit(chargeLimitPercent int32) error {
	return conn.runCommand(func(ctx context.Context, car *vehicle.Vehicle) error {
		return car.ChangeChargeLimit(ctx, chargeLimitPercent)
	})
}
