package conf

type Config struct {
	PublicKey  string   `json:"public_key"`
	PrivateKey string   `json:"private_key"`
	ProjectId  string   `json:"project_id"`
	Region     string   `json:"region"`
	Migrate    *Migrate `json:"migrate" validate:"required"`
	Log        *Log     `json:"log"`
}

type Log struct {
	IsStdout bool   `json:"is_stdout"`
	Dir      string `json:"dir"`
	Name     string `json:"name"`
	Level    string `json:"level"`
}

type Migrate struct {
	BasicConfig
	ServiceValidation *ServiceValidation `json:"service_validation"`
}

type ServiceValidation struct {
	Port                    int `json:"port" validate:"required"`
	WaitServiceReadyTimeout int `json:"wait_service_ready_timeout" validate:"required"`
}

type BasicConfig struct {
	UHostConfig *UHostConfig `json:"uhost_config" validate:"required"`
	CubeConfig  *CubeConfig  `json:"cube_config"`
}

type CubeConfig struct {
	CubeIdList   []string      `json:"cube_id_list"`
	CubeIdFilter *CubeIdFilter `json:"cube_id_filter"`
}

type UHostConfig struct {
	Zone               string         `json:"zone" validate:"required"`
	ImageId            string         `json:"image_id"`
	ImageIdFilter      *ImageIdFilter `json:"image_id_filter"`
	Name               string         `json:"name"`
	NamePrefix         string         `json:"name_prefix"`
	Password           string         `json:"password" validate:"required"`
	ChargeType         string         `json:"charge_type"`
	Duration           int            `json:"duration"`
	CPU                int            `json:"cpu"`
	Memory             int            `json:"memory"`
	Tag                string         `json:"tag"`
	MinimalCpuPlatform string         `json:"minimal_cpu_platform"`
	MachineType        string         `json:"machine_type"`
	VPCId              string         `json:"vpc_id"`
	SubnetId           string         `json:"subnet_id"`
	SecurityGroupId    string         `json:"security_group_id"`
	Disks              []UHostDisk    `json:"disks" validate:"required"`
}

type UHostDisk struct {
	IsBoot string `json:"is_boot" validate:"required"`
	Size   int    `json:"size" validate:"required"`
	Type   string `json:"type" validate:"required"`
}

type CubeIdFilter struct {
	Zone         string `json:"zone"`
	VPCId        string `json:"vpc_id"`
	SubnetId     string `json:"subnet_id"`
	Group        string `json:"group"`
	DeploymentId string `json:"deployment_id"`
	NameRegex    string `json:"name_regex"`
}

type ImageIdFilter struct {
	Zone       string `json:"zone"`
	OSType     string `json:"os_type"`
	ImageType  string `json:"image_type"`
	MostRecent bool   `json:"most_recent"`
	NameRegex  string `json:"name_regex"`
}
