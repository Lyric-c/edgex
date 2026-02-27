package model

import (
	"encoding/json"
	"errors"
	"time"
)

type Duration time.Duration

// RegisterType 定义寄存器类型
type RegisterType int

const (
	RegHolding       RegisterType = iota // 0: 默认Holding寄存器 (03)
	RegCoil                              // 1: Coil (01)
	RegDiscreteInput                     // 2: Discrete Input (02)
	RegInput                             // 3: Input Register (04)
	RegCustom                            // 4: 非标准
)

func (r RegisterType) FunctionCode() byte {
	switch r {
	case RegCoil:
		return 1
	case RegDiscreteInput:
		return 2
	case RegHolding:
		return 3
	case RegInput:
		return 4
	case RegCustom:
		return 0 // 由配置指定
	}
	return 0
}

func (r RegisterType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r RegisterType) MarshalYAML() (interface{}, error) {
	return r.String(), nil
}

func (r RegisterType) String() string {
	switch r {
	case RegCoil:
		return "Coils (outputs)"
	case RegDiscreteInput:
		return "Discrete Inputs"
	case RegHolding:
		return "Holding Registers"
	case RegInput:
		return "Input Registers"
	case RegCustom:
		return "Custom"
	}
	return "Unknown"
}

// Code 返回Modbus功能码
func (r RegisterType) Code() byte {
	switch r {
	case RegCoil:
		return 1
	case RegDiscreteInput:
		return 2
	case RegHolding:
		return 3
	case RegInput:
		return 4
	case RegCustom:
		return 0
	}
	return 0
}

// ShortString 返回短名称
func (r RegisterType) ShortString() string {
	switch r {
	case RegCoil:
		return "coil"
	case RegDiscreteInput:
		return "discrete_input"
	case RegHolding:
		return "holding"
	case RegInput:
		return "input"
	case RegCustom:
		return "custom"
	}
	return "unknown"
}

func (r *RegisterType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var n int
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		*r = RegisterType(n)
		return nil
	}
	*r = ParseRegisterType(s)
	return nil
}

func (r *RegisterType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		var n int
		if err := unmarshal(&n); err != nil {
			return err
		}
		*r = RegisterType(n)
		return nil
	}
	*r = ParseRegisterType(s)
	return nil
}

func ParseRegisterType(s string) RegisterType {
	switch s {
	case "coil", "1", "COIL", "Coil", "Coils (outputs)", "Coils":
		return RegCoil
	case "discrete_input", "2", "DISCRETE_INPUT", "DiscreteInput", "discrete", "Discrete Inputs":
		return RegDiscreteInput
	case "holding", "3", "HOLDING", "Holding", "holding_register", "4x", "Holding Registers", "HoldingRegister":
		return RegHolding
	case "input", "4", "INPUT", "Input", "input_register", "3x", "Input Registers", "InputRegister":
		return RegInput
	case "custom", "0", "Custom":
		return RegCustom
	}
	return RegHolding
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	val, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(val)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
	default:
		return errors.New("invalid duration")
	}
	return nil
}

// Point represents a data point configuration (Tag/Variable)
type Point struct {
	ID           string           `json:"id" yaml:"id"`
	Name         string           `json:"name" yaml:"name"`
	RegisterType RegisterType     `json:"register_type" yaml:"register_type"`
	FunctionCode byte             `json:"function_code" yaml:"function_code"` // 允许非标准功能码 (当前会将配置初始化值设置为0)
	Address      string           `json:"address" yaml:"address"`             // 地址字符串，支持不同协议格式
	DataType     string           `json:"datatype" yaml:"datatype"`           // int16, float32, bool, bit.0
	Scale        float64          `json:"scale" yaml:"scale"`
	Offset       float64          `json:"offset" yaml:"offset"`
	Format       string           `json:"format,omitempty" yaml:"format,omitempty"`
	WordOrder    string           `json:"word_order,omitempty" yaml:"word_order,omitempty"`
	ReadFormula  string           `json:"read_formula,omitempty" yaml:"read_formula,omitempty"`
	WriteFormula string           `json:"write_formula,omitempty" yaml:"write_formula,omitempty"`
	Unit         string           `json:"unit" yaml:"unit"`
	ReadWrite    string           `json:"readwrite" yaml:"readwrite"` // R / RW
	Group        string           `json:"group" yaml:"group"`
	ReportMode   string           `json:"report_mode" yaml:"report_mode"` // cycle / cov / event
	Threshold    *ThresholdConfig `json:"threshold" yaml:"threshold"`
	DeviceID     string           `json:"-" yaml:"-"` // Runtime field, not persisted
}

// ThresholdConfig defines alarm thresholds for a point
type ThresholdConfig struct {
	High float64 `json:"high" yaml:"high"`
	Low  float64 `json:"low" yaml:"low"`
}

// Value represents the standardized output of a collected point
type Value struct {
	ChannelID string    `json:"channel_id"`
	DeviceID  string    `json:"device_id"`
	PointID   string    `json:"point_id"`
	Value     any       `json:"value"`
	Quality   string    `json:"quality"`
	TS        time.Time `json:"timestamp"`
}

// PointData represents point configuration and current value for frontend display
type PointData struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SlaveID      uint8     `json:"slave_id"`
	RegisterType string    `json:"register_type"`
	FunctionCode byte      `json:"function_code"`
	Address      string    `json:"address"`
	DataType     string    `json:"datatype"`
	Value        any       `json:"value"`
	Quality      string    `json:"quality"`
	Timestamp    time.Time `json:"timestamp"`
	Unit         string    `json:"unit,omitempty"`
	ReadWrite    string    `json:"readwrite"` // R / RW
}

// DeviceStorage defines data storage strategy for a device
type DeviceStorage struct {
	Enable     bool   `json:"enable" yaml:"enable"`
	Strategy   string `json:"strategy" yaml:"strategy"`       // "realtime" (every record), "interval" (fixed time)
	Interval   int    `json:"interval" yaml:"interval"`       // Storage interval in minutes (1, 10, 60)
	MaxRecords int    `json:"max_records" yaml:"max_records"` // Max history records (default 1000)
}

// Device represents a device configuration (within a channel)
type Device struct {
	ID         string         `json:"id" yaml:"id"`
	Name       string         `json:"name" yaml:"name"`
	Enable     bool           `json:"enable" yaml:"enable"`
	Interval   Duration       `json:"interval" yaml:"interval"`
	DeviceFile string         `json:"device_file,omitempty" yaml:"device_file,omitempty"` // 设备配置文件路径
	Config     map[string]any `json:"config" yaml:"config"`                               // 设备特定配置（如 slave_id）
	Storage    DeviceStorage  `json:"storage,omitempty" yaml:"storage,omitempty"`         // Data storage strategy
	Points     []Point        `json:"points" yaml:"points"`                               // 该设备的点位列表
	State      int            `json:"state" yaml:"-"`                                     // 运行时状态：0=Online, 1=Unstable, 2=Offline, 3=Quarantine
	StopChan   chan struct{}  `json:"-" yaml:"-"`
	// Runtime state fields
	NodeRuntime *NodeRuntime `json:"runtime,omitempty" yaml:"-"`
}

// NodeRuntime defines runtime statistics for a node (device or channel)
type NodeRuntime struct {
	FailCount     int       `json:"fail_count"`
	SuccessCount  int       `json:"success_count"`
	LastFailTime  time.Time `json:"last_fail_time"`
	NextRetryTime time.Time `json:"next_retry_time"`
	State         int       `json:"state"` // NodeState enum
}

// Channel represents a collection channel (采集通道)
// 一个通道对应一个采集驱动 (如 Modbus TCP, S7, Modbus RTU 等)
type Channel struct {
	ID       string         `json:"id" yaml:"id"`
	Name     string         `json:"name" yaml:"name"`
	Protocol string         `json:"protocol" yaml:"protocol"` // modbus-tcp, modbus-rtu, s7, opc-ua, etc.
	Enable   bool           `json:"enable" yaml:"enable"`
	Config   map[string]any `json:"config" yaml:"config"`   // 协议特定配置 (IP, Port, etc.)
	Devices  []Device       `json:"devices" yaml:"devices"` // 该通道下的设备列表
	StopChan chan struct{}  `json:"-" yaml:"-"`
	// Runtime fields
	NodeRuntime *NodeRuntime `json:"runtime,omitempty" yaml:"-"`
}

// DriverConfig is the configuration passed to a driver
type DriverConfig struct {
	ChannelID string         `json:"channel_id"`
	Protocol  string         `json:"protocol"` // Protocol name (e.g. modbus-tcp, modbus-rtu)
	Config    map[string]any `json:"config"`
}

// NorthboundConfig defines configuration for northbound data reporting
type NorthboundConfig struct {
	MQTT       []MQTTConfig       `json:"mqtt" yaml:"mqtt"`
	HTTP       []HTTPConfig       `json:"http" yaml:"http"`
	OPCUA      []OPCUAConfig      `json:"opcua" yaml:"opcua"`
	SparkplugB []SparkplugBConfig `json:"sparkplug_b" yaml:"sparkplug_b"`
	Status     map[string]int     `json:"status,omitempty" yaml:"-"`
}

type DataCacheConfig struct {
	Enable        bool   `json:"enable" yaml:"enable"`
	MaxCount      int    `json:"max_count" yaml:"max_count"`           // Default 1000
	FlushInterval string `json:"flush_interval" yaml:"flush_interval"` // e.g. "1m"
}

type MQTTConfig struct {
	ID             string `json:"id" yaml:"id"`
	Name           string `json:"name" yaml:"name"`
	Enable         bool   `json:"enable" yaml:"enable"`
	Broker         string `json:"broker" yaml:"broker"`
	ClientID       string `json:"client_id" yaml:"client_id"`
	Topic          string `json:"topic" yaml:"topic"`
	SubscribeTopic string `json:"subscribe_topic" yaml:"subscribe_topic"` // New: Subscribe topic for write requests

	StatusTopic          string `json:"status_topic" yaml:"status_topic"`                     // Online/Offline status topic
	LwtTopic             string `json:"lwt_topic" yaml:"lwt_topic"`                           // LWT topic (if different from StatusTopic)
	DeviceStatusTopic    string `json:"device_status_topic" yaml:"device_status_topic"`       // Sub-device Online/Offline topic
	DeviceLifecycleTopic string `json:"device_lifecycle_topic" yaml:"device_lifecycle_topic"` // Sub-device Add/Remove topic
	OnlinePayload        string `json:"online_payload" yaml:"online_payload"`                 // Payload for online status
	OfflinePayload       string `json:"offline_payload" yaml:"offline_payload"`               // Payload for offline status (graceful disconnect)
	LwtPayload           string `json:"lwt_payload" yaml:"lwt_payload"`                       // Payload for LWT (ungraceful disconnect)
	IgnoreOfflineData    bool   `json:"ignore_offline_data" yaml:"ignore_offline_data"`       // If true, do not report data when device is offline

	WriteResponseTopic string `json:"write_response_topic" yaml:"write_response_topic"` // Topic for write responses

	Username string                         `json:"username" yaml:"username"`
	Password string                         `json:"password" yaml:"password"`
	Cache    DataCacheConfig                `json:"cache" yaml:"cache"`
	Devices  map[string]DevicePublishConfig `json:"devices" yaml:"devices"`
}

type HTTPConfig struct {
	ID                  string            `json:"id" yaml:"id"`
	Name                string            `json:"name" yaml:"name"`
	Enable              bool              `json:"enable" yaml:"enable"`
	URL                 string            `json:"url" yaml:"url"`       // Base URL
	Method              string            `json:"method" yaml:"method"` // POST/PUT
	Headers             map[string]string `json:"headers" yaml:"headers"`
	AuthType            string            `json:"auth_type" yaml:"auth_type"` // None, Basic, Bearer, APIKey
	Username            string            `json:"username" yaml:"username"`
	Password            string            `json:"password" yaml:"password"`
	Token               string            `json:"token" yaml:"token"`
	APIKeyName          string            `json:"api_key_name" yaml:"api_key_name"`
	APIKeyValue         string            `json:"api_key_value" yaml:"api_key_value"`
	DataEndpoint        string            `json:"data_endpoint" yaml:"data_endpoint"`                 // Relative path for data
	DeviceEventEndpoint string            `json:"device_event_endpoint" yaml:"device_event_endpoint"` // Relative path for events
	Cache               DataCacheConfig   `json:"cache" yaml:"cache"`
	Devices             map[string]bool   `json:"devices" yaml:"devices"` // Key: DeviceID, Value: Enable
}

type DevicePublishConfig struct {
	Enable   bool     `json:"enable" yaml:"enable"`
	Strategy string   `json:"strategy" yaml:"strategy"` // "periodic" or "cov"
	Interval Duration `json:"interval" yaml:"interval"` // 0 means use collection interval
}

type OPCUAConfig struct {
	ID              string            `json:"id" yaml:"id"`
	Name            string            `json:"name" yaml:"name"`
	Enable          bool              `json:"enable" yaml:"enable"`
	Port            int               `json:"port" yaml:"port"`
	Endpoint        string            `json:"endpoint" yaml:"endpoint"`
	SecurityPolicy  string            `json:"security_policy" yaml:"security_policy"` // "None", "Basic256", "Basic256Sha256", "Auto"
	SecurityMode    string            `json:"security_mode" yaml:"security_mode"`     // "None", "Sign", "SignAndEncrypt"
	TrustedCertPath string            `json:"trusted_cert_path" yaml:"trusted_cert_path"`
	AuthMethods     []string          `json:"auth_methods" yaml:"auth_methods"` // "Anonymous", "UserName", "Certificate"
	Users           map[string]string `json:"users" yaml:"users"`               // Username -> Password
	CertFile        string            `json:"cert_file" yaml:"cert_file"`       // Path to server certificate
	KeyFile         string            `json:"key_file" yaml:"key_file"`         // Path to server private key
	Devices         map[string]bool   `json:"devices" yaml:"devices"`           // Key: DeviceID, Value: Enable
}

type SparkplugBConfig struct {
	ID             string          `json:"id" yaml:"id"`
	Name           string          `json:"name" yaml:"name"`
	Enable         bool            `json:"enable" yaml:"enable"`
	ClientID       string          `json:"client_id" yaml:"client_id"`
	GroupID        string          `json:"group_id" yaml:"group_id"`
	NodeID         string          `json:"node_id" yaml:"node_id"`
	EnableAlias    bool            `json:"enable_alias" yaml:"enable_alias"`
	GroupPath      bool            `json:"group_path" yaml:"group_path"`
	OfflineCache   bool            `json:"offline_cache" yaml:"offline_cache"`
	CacheMemSize   int             `json:"cache_mem_size" yaml:"cache_mem_size"`
	CacheDiskSize  int             `json:"cache_disk_size" yaml:"cache_disk_size"`
	CacheResendInt int             `json:"cache_resend_int" yaml:"cache_resend_int"`
	Broker         string          `json:"broker" yaml:"broker"`
	Port           int             `json:"port" yaml:"port"`
	Username       string          `json:"username" yaml:"username"`
	Password       string          `json:"password" yaml:"password"`
	SSL            bool            `json:"ssl" yaml:"ssl"`
	CACert         string          `json:"ca_cert" yaml:"ca_cert"`
	ClientCert     string          `json:"client_cert" yaml:"client_cert"`
	ClientKey      string          `json:"client_key" yaml:"client_key"`
	KeyPassword    string          `json:"key_password" yaml:"key_password"`
	Devices        map[string]bool `json:"devices" yaml:"devices"` // Key: DeviceID, Value: Enable
}

// EdgeRule represents an edge computing rule
type EdgeRule struct {
	ID            string        `json:"id" yaml:"id"`
	Name          string        `json:"name" yaml:"name"`
	Type          string        `json:"type" yaml:"type"` // threshold, calculation, state, window
	Enable        bool          `json:"enable" yaml:"enable"`
	Priority      int           `json:"priority" yaml:"priority"`
	CheckInterval string        `json:"check_interval" yaml:"check_interval"` // e.g. "5s", "1m"
	TriggerMode   string        `json:"trigger_mode" yaml:"trigger_mode"`     // always, on_change
	Source        RuleSource    `json:"source" yaml:"source"`                 // Deprecated: use Sources
	Sources       []RuleSource  `json:"sources" yaml:"sources"`               // New: Multiple sources
	TriggerLogic  string        `json:"trigger_logic" yaml:"trigger_logic"`   // "AND", "OR", "EXPR"
	Condition     string        `json:"condition" yaml:"condition"`           // Boolean Expression
	Expression    string        `json:"expression" yaml:"expression"`         // Calculation Expression
	Actions       []RuleAction  `json:"actions" yaml:"actions"`
	Window        *WindowConfig `json:"window,omitempty" yaml:"window,omitempty"`
	State         *StateConfig  `json:"state,omitempty" yaml:"state,omitempty"`
}

type RuleSource struct {
	Alias     string `json:"alias" yaml:"alias"` // Variable name in expression (e.g. "t1")
	ChannelID string `json:"channel_id" yaml:"channel_id"`
	DeviceID  string `json:"device_id" yaml:"device_id"`
	PointID   string `json:"point_id" yaml:"point_id"`
	PointName string `json:"point_name" yaml:"point_name"`
}

type RuleAction struct {
	Type   string         `json:"type" yaml:"type"` // mqtt, http, log, command
	Config map[string]any `json:"config" yaml:"config"`
}

type WindowConfig struct {
	Type     string `json:"type" yaml:"type"`           // sliding, tumbling
	Size     string `json:"size" yaml:"size"`           // e.g. "10s", "100" (count)
	Interval string `json:"interval" yaml:"interval"`   // Step size for sliding
	AggrFunc string `json:"aggr_func" yaml:"aggr_func"` // avg, min, max, sum, count
}

type StateConfig struct {
	Duration string `json:"duration" yaml:"duration"` // e.g. "10s" (Hold time)
	Count    int    `json:"count" yaml:"count"`       // Consecutive count
}

// RuleRuntimeState represents the runtime status of a rule
type RuleRuntimeState struct {
	RuleID         string            `json:"rule_id"`
	RuleName       string            `json:"rule_name"`
	Enable         bool              `json:"enable"`
	LastCheckTime  time.Time         `json:"last_check_time,omitempty"` // For CheckInterval
	LastTrigger    time.Time         `json:"last_trigger"`
	LastValue      any               `json:"last_value"`
	TriggerCount   int64             `json:"trigger_count"`
	CurrentStatus  string            `json:"current_status"` // NORMAL, ALARM
	ConditionStart time.Time         `json:"condition_start,omitempty"`
	ConditionCount int               `json:"condition_count,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	ActionLastRuns map[int]time.Time `json:"action_last_runs,omitempty"`
}

type FailedAction struct {
	ID         string         `json:"id"`
	RuleID     string         `json:"rule_id"`
	Action     RuleAction     `json:"action"`
	Value      Value          `json:"value"`
	Timestamp  time.Time      `json:"timestamp"`
	RetryCount int            `json:"retry_count"`
	LastError  string         `json:"last_error"`
	Env        map[string]any `json:"env"`
}

// SouthboundManager interface defines methods required by Northbound components
// to interact with southbound devices (e.g. for building address space or writing values)
type SouthboundManager interface {
	GetChannels() []Channel
	GetChannelDevices(channelID string) []Device
	GetDevice(channelID, deviceID string) *Device
	WritePoint(channelID, deviceID, pointID string, value any) error
}
