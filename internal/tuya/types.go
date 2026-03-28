package tuya

type TokenResponse struct {
	Success bool        `json:"success"`
	Result  TokenResult `json:"result"`
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
}

type TokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireTime   int    `json:"expire_time"`
	UID          string `json:"uid"`
}

type DeviceStatusResponse struct {
	Success bool         `json:"success"`
	Result  []StatusItem `json:"result"`
	Code    int          `json:"code"`
	Msg     string       `json:"msg"`
}

type StatusItem struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

type DeviceLogsResponse struct {
	Success bool             `json:"success"`
	Result  DeviceLogsResult `json:"result"`
	Code    int              `json:"code"`
	Msg     string           `json:"msg"`
}

type DeviceLogsResult struct {
	Logs          []LogItem `json:"logs"`
	HasNext       bool      `json:"has_next"`
	NextRowKey    string    `json:"next_row_key"`
	CurrentRowKey string    `json:"current_row_key"`
}

type LogItem struct {
	Code      string      `json:"code"`
	Value     interface{} `json:"value"`
	EventTime int64       `json:"event_time"`
	EventFrom string      `json:"event_from"`
	EventID   int         `json:"event_id"`
}

type DeviceInfoResponse struct {
	Success bool       `json:"success"`
	Result  DeviceInfo `json:"result"`
	Code    int        `json:"code"`
	Msg     string     `json:"msg"`
}

type DeviceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
