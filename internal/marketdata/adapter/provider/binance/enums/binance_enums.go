package binance

type SymbolStatus string

const (
	SymbolStatusTrading   SymbolStatus = "TRADING"
	SymbolStatusEndOfDay  SymbolStatus = "END_OF_DAY"
	SymbolStatusHalt      SymbolStatus = "HALT"
	SymbolStatusBreak     SymbolStatus = "BREAK"
)

func (s SymbolStatus) IsValid() bool {
	switch s {
	case
		SymbolStatusTrading,
		SymbolStatusEndOfDay,
		SymbolStatusHalt,
		SymbolStatusBreak:
		return true
	}
	return false
}

type Permission string

const (
	PermissionSpot      Permission = "SPOT"
	PermissionMargin    Permission = "MARGIN"
	PermissionLeveraged Permission = "LEVERAGED"
)

func (p Permission) IsValid() bool {

	switch p {
	case
		PermissionSpot,
		PermissionMargin,
		PermissionLeveraged:
		return true
	}

	if len(p) >= 7 && p[:7] == "TRD_GRP" {
		return true
	}

	return false
}


type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPendingNew      OrderStatus = "PENDING_NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
	OrderStatusExpiredInMatch  OrderStatus = "EXPIRED_IN_MATCH"
)

func (s OrderStatus) IsFinal() bool {

	switch s {
	case
		OrderStatusFilled,
		OrderStatusCanceled,
		OrderStatusRejected,
		OrderStatusExpired,
		OrderStatusExpiredInMatch:
		return true
	}

	return false
}

type OrderType string

const (
	OrderTypeLimit           OrderType = "LIMIT"
	OrderTypeMarket          OrderType = "MARKET"
	OrderTypeStopLoss        OrderType = "STOP_LOSS"
	OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"
	OrderTypeLimitMaker      OrderType = "LIMIT_MAKER"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceIOC TimeInForce = "IOC"
	TimeInForceFOK TimeInForce = "FOK"
)

type OrderResponseType string

const (
	OrderRespACK    OrderResponseType = "ACK"
	OrderRespRESULT OrderResponseType = "RESULT"
	OrderRespFULL   OrderResponseType = "FULL"
)

type RateLimitType string

const (
	RateLimitRequestWeight RateLimitType = "REQUEST_WEIGHT"
	RateLimitOrders        RateLimitType = "ORDERS"
	RateLimitRawRequests   RateLimitType = "RAW_REQUESTS"
)


type RateLimitInterval string

const (
	IntervalSecond RateLimitInterval = "SECOND"
	IntervalMinute RateLimitInterval = "MINUTE"
	IntervalDay    RateLimitInterval = "DAY"
)

type STPMode string

const (
	STPNone        STPMode = "NONE"
	STPExpireMaker STPMode = "EXPIRE_MAKER"
	STPExpireTaker STPMode = "EXPIRE_TAKER"
	STPExpireBoth  STPMode = "EXPIRE_BOTH"
	STPDecrement   STPMode = "DECREMENT"
	STPTransfer    STPMode = "TRANSFER"
)

type ContingencyType string

const (
	ContingencyOCO ContingencyType = "OCO"
	ContingencyOTO ContingencyType = "OTO"
)

type AllocationType string

const (
	AllocationSOR AllocationType = "SOR"
)


type ListStatusType string

const (
	ListStatusResponse    ListStatusType = "RESPONSE"
	ListStatusExecStarted ListStatusType = "EXEC_STARTED"
	ListStatusUpdated     ListStatusType = "UPDATED"
	ListStatusAllDone     ListStatusType = "ALL_DONE"
)

type ListOrderStatus string

const (
	ListOrderStatusExecuting ListOrderStatus = "EXECUTING"
	ListOrderStatusAllDone   ListOrderStatus = "ALL_DONE"
	ListOrderStatusReject    ListOrderStatus = "REJECT"
)


type WorkingFloor string

const (
	WorkingFloorExchange WorkingFloor = "EXCHANGE"
	WorkingFloorSOR      WorkingFloor = "SOR"
)


type RateLimit struct {

	Type        RateLimitType     `json:"rateLimitType"`

	Interval    RateLimitInterval `json:"interval"`

	IntervalNum int               `json:"intervalNum"`

	Limit       int               `json:"limit"`
}

func ParseOrderSide(v string) (OrderSide, error) {

	side := OrderSide(v)

	switch side {

	case OrderSideBuy, OrderSideSell:
		return side, nil
	}

	return "", fmt.Errorf("invalid OrderSide: %s", v)
}
