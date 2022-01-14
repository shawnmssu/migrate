// Code is generated by ucloud-model, DO NOT EDIT IT.

package cube

/*
EIPAddr - EIP地址
*/
type EIPAddr struct {

	// IP地址
	IP string

	// 线路名称BGP或者internalation
	OperatorName string
}

/*
EIPSet - EIP信息
*/
type EIPSet struct {

	// EIP带宽值
	Bandwidth int

	// 带宽类型0标准普通带宽，1表示共享带宽
	BandwidthType int

	// EIP创建时间
	CreateTime int

	// EIP地址
	EIPAddr []EIPAddr

	// EIPId
	EIPId string

	// 付费模式，带宽付费或者流量付费
	PayMode string

	// EIP绑定对象的资源Id
	Resource string

	// EIP状态，表示使用中或者空闲
	Status string

	// EIP权重
	Weight int
}

/*
CubeExtendInfo - Cube的额外信息
*/
type CubeExtendInfo struct {

	// Cube的Id
	CubeId string

	// EIPSet
	Eip []EIPSet

	// 资源有效期
	Expiration int

	// Cube的名称
	Name string

	// 业务组名称
	Tag string
}

/*
ValueSet -
*/
type ValueSet struct {

	//
	Timestamp int

	//
	Value float64
}

/*
MetricDataSet - 监控数据集合
*/
type MetricDataSet struct {

	//
	MetricName string

	//
	Values []ValueSet
}