package tesla

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/teslamotors/vehicle-command/pkg/cache"
	"github.com/teslamotors/vehicle-command/pkg/connector/ble"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
	"github.com/teslamotors/vehicle-command/pkg/vehicle"
)

const commandTimeout = 30 * time.Second

type Connection struct {
	privKey      protocol.ECDHPrivateKey
	vin          string
	sessionCache *cache.SessionCache
	commandLock  sync.Mutex
}

func MakeConnection(privateKey []byte, vin string) (*Connection, error) {
	privKey, err := loadPrivKey(privateKey)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		privKey:      privKey,
		vin:          vin,
		sessionCache: cache.New(0),
	}

	return conn, nil
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

// Copy/Paste from LoadExternalECDHKey
var ErrInvalidPrivateKey = errors.New("invalid private key")

func loadPrivKey(pemBlock []byte) (protocol.ECDHPrivateKey, error) {
	block, _ := pem.Decode([]byte(pemBlock))
	if block == nil {
		return nil, fmt.Errorf("%w: expected PEM encoding", ErrInvalidPrivateKey)
	}

	var ecdsaPrivateKey *ecdsa.PrivateKey
	var err error

	if block.Type == "EC PRIVATE KEY" {
		ecdsaPrivateKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else {
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		if ecdsaPrivateKey, ok = privateKey.(*ecdsa.PrivateKey); !ok {
			return nil, fmt.Errorf("%w: only elliptic curve keys supported", ErrInvalidPrivateKey)
		}
	}

	if ecdsaPrivateKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("%w: only NIST-P256 keys supported", ErrInvalidPrivateKey)
	}

	privateScalar := ecdsaPrivateKey.D.Bytes()

	return protocol.UnmarshalECDHPrivateKey(privateScalar), nil
}
