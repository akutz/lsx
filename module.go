package lsx

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
)

// Module is the interface that defines the basis for this program's
// modular components.
type Module interface {
	// Name returns the name of the module.
	Name() string

	// Type returns the type of the module.
	Type() string

	// Init initializes the module.
	Init(ctx context.Context) error
}

var (
	mods    = map[ModuleType]map[string]func() Module{}
	modsRWL = sync.RWMutex{}
)

// ModuleType is used to define constant module types.
type ModuleType uint8

const (
	// InvalidModuleType is an invalid module type.
	InvalidModuleType ModuleType = iota

	// ClientModuleType is a module that provides an implementation of the
	// Client interface.
	ClientModuleType

	// ConfigModuleType is a module that provides an implementation of the
	// Config interface.
	ConfigModuleType

	// LoggerModuleType is a module that provides an implementation of the
	// Logger interface.
	LoggerModuleType

	// ServerModuleType is a module that provides an implementation of the
	// Server interface.
	ServerModuleType

	// VolumeModuleType is a module that provides an implementation of the
	// VolumeDriver interface.
	VolumeModuleType
)

const (
	// maxModuleType is the max, valid module type. Used for iterating the
	// module type constants.
	maxModuleType = VolumeModuleType
)

// String returns the module type's string representation.
func (t ModuleType) String() string {
	switch t {
	case ClientModuleType:
		return "client"
	case ConfigModuleType:
		return "config"
	case LoggerModuleType:
		return "logger"
	case ServerModuleType:
		return "server"
	case VolumeModuleType:
		return "volume"
	}
	return "invalid"
}

// ParseModuleType parses a numeric, string, or ModuleType value and
// returns the corresponding module type.
func ParseModuleType(v interface{}) (ModuleType, error) {
	switch tv := v.(type) {
	case *ModuleType:
		return *tv, nil
	case ModuleType:
		return isValidModuleType(tv)
	case *uint8:
		return ParseModuleType(*tv)
	case uint8:
		return isValidModuleType(ModuleType(tv))
	case *uint16:
		return ParseModuleType(*tv)
	case uint16:
		if tv <= 255 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *uint32:
		return ParseModuleType(*tv)
	case uint32:
		if tv <= 255 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *uint:
		return ParseModuleType(*tv)
	case uint:
		return ParseModuleType(uint64(tv))
	case *uint64:
		return ParseModuleType(*tv)
	case uint64:
		if tv <= 255 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *int8:
		return ParseModuleType(*tv)
	case int8:
		if tv > 0 && tv <= 127 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *int16:
		return ParseModuleType(*tv)
	case int16:
		if tv > 0 && tv <= 255 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *int32:
		return ParseModuleType(*tv)
	case int32:
		if tv > 0 && tv <= 255 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *int:
		return ParseModuleType(*tv)
	case int:
		return ParseModuleType(int64(tv))
	case *int64:
		return ParseModuleType(*tv)
	case int64:
		if tv > 0 && tv <= 255 {
			return isValidModuleType(ModuleType(uint8(tv)))
		}
	case *float32:
		return ParseModuleType(*tv)
	case float32:
		if tv > 0 && tv <= 255 {
			if isWholeInt := math.Mod(float64(tv), 1) == 0; isWholeInt {
				return isValidModuleType(ModuleType(uint8(tv)))
			}
		}
	case *float64:
		return ParseModuleType(*tv)
	case float64:
		if tv > 0 && tv <= 255 {
			if isWholeInt := math.Mod(tv, 1) == 0; isWholeInt {
				return isValidModuleType(ModuleType(uint8(tv)))
			}
		}
	case fmt.Stringer:
		return ParseModuleType(tv.String())
	case *string:
		return ParseModuleType(*tv)
	case string:
		for mt := InvalidModuleType + 1; mt <= maxModuleType; mt++ {
			if strings.EqualFold(tv, mt.String()) {
				return mt, nil
			}
		}
	}
	return errUnknownModuleType(v)
}

func isValidModuleType(v ModuleType) (ModuleType, error) {
	if v < InvalidModuleType+1 || v > maxModuleType {
		return errUnknownModuleType(v)
	}
	return v, nil
}

func errUnknownModuleType(v interface{}) (ModuleType, error) {
	if mt, ok := v.(ModuleType); ok {
		v = uint8(mt)
	}
	return 0, fmt.Errorf("error: invalid module type: %v", v)
}

// RegisterModule a new module.
func RegisterModule(modType ModuleType, modName string, modCtor func() Module) {
	modsRWL.Lock()
	defer modsRWL.Unlock()
	modCtorMap, ok := mods[modType]
	if !ok {
		modCtorMap = map[string]func() Module{}
		mods[modType] = modCtorMap
	}
	modCtorMap[modName] = modCtor
}

// NewModule returns a new instance of a registered module type.
func NewModule(modType ModuleType, modName string) Module {
	serverCtorsRWL.RLock()
	defer serverCtorsRWL.RUnlock()
	if a, ok := mods[modType]; ok {
		if b, ok := a[modName]; ok {
			return b()
		}
	}
	return nil
}
