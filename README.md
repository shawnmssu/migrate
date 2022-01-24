# Migrate

Migrate CLI provides a unified command line interface for migrating Cube instances to UHost instances at the present. 

The Process about:
- `migrate eip`
  - Create UHost List (Default: Shared Outstanding UHost).
  - Waiting for UHost Running.
  - Unbind one of the queried Cube with EIP and then bind EIP to UHost.
  - Repeat the previous step.
  - [Option]Running tcp validation about UHost service.
  - [Option]RollBack the EIP to Cube When migrate got error.
- `migrate ulb`
  - Create UHost List (Default: Shared Outstanding UHost).
  - Waiting for UHost Running.
  - Create one ULB VServer Backend used UHost and then delete one ulb backend about Cube.(migrate ulb policy not supported yet).
  - Repeat the previous step.
  - [Option]Running ulb backend health check.
  - [Option]RollBack the ulb backend about Cube When migrate got error.
- `migrate private-ip`
  - Get Cube list which need to migrate by `cube_config`.
  - Delete one Cube for freed the private ip.
  - Create UHost By the private ip and UHost Config (Default: Shared Outstanding UHost).
  - Waiting for UHost Running.
  - Repeat the previous step.
  - [Option/Recommend] use the `dry-run` flag to validate UHost Config and Cube config.
    - if set true which means that only try to create one temporary UHost with "Dynamic" `charge_type` to validate `uhost_config`, waiting for UHost running and then delete it.
    - for example:`migrate private-ip --conf xxx --dry-run` 

## Installation

**You can install by [release](https://github.com/shawnmssu/migrate/releases)**
- download release
- decompress release, for example:
```shell
tar zxvf migrate_x.x.x_linux-amd64.tgz
```

**Building from source(Recommended if you have golang installed)**
- Go 1.16
- You have installed git and golang on your platform, you can fetch the source code from GitHub and compile it by yourself.

```bash
git clone https://github.com/shawnmssu/migrate.git
cd migrate
make install
migrate --help
```

## Use

```
migrate --conf configs/config.json
```

## Config example
- `migrate eip` for example:
```json
{
  "public_key": "xxx",
  "private_key": "xxx",
  "project_id": "org-xxx",
  "region": "hk",
  "migrate_eip": {
    "uhost_config": {
      "zone":  "hk-02",
      "image_id_filter":  {
        "os_type": "Linux",
        "image_type": "Base",
        "most_recent": true
      },
      "name_prefix":  "Test-migrate",
      "password": "xxx",
      "charge_type":  "Month",
      "cpu":  1,
      "memory": 1024,
      "tag":  "migrate",
      "minimal_cpu_platform": "Amd/Auto",
      "machine_type":  "OM",
      "disks": [
        {
          "is_boot": "True",
          "size": 20,
          "type": "CLOUD_RSSD"
        }
      ]
    },
    "cube_config": {
      "cube_id_filter": {
        "zone": "hk-02"
      }
    },
    "service_validation": {
      "port": 22,
      "wait_service_ready_timeout": 120
    }
  }
}
```
- `migrate ulb` for example:
```json
{
  "public_key": "xxx",
  "private_key": "xxx",
  "project_id": "org-xxx",
  "region": "hk",
  "migrate_ulb": {
    "ulb_id": "ulb-xxx",
    "uhost_config": {
      "zone":  "hk-02",
      "image_id_filter":  {
        "os_type": "Linux",
        "image_type": "Base",
        "most_recent": true
      },
      "name_prefix":  "Test-migrate",
      "password": "xxx",
      "charge_type":  "Month",
      "cpu":  1,
      "memory": 1024,
      "tag":  "migrate",
      "minimal_cpu_platform": "Amd/Auto",
      "machine_type":  "OM",
      "disks": [
        {
          "is_boot": "True",
          "size": 20,
          "type": "CLOUD_RSSD"
        }
      ]
    }
  }
}
```
- `migrate private-ip` for example:
```json
{
  "public_key": "xxx",
  "private_key": "xxx",
  "project_id": "org-xxx",
  "region": "hk",
  "migrate_private_ip": {
    "uhost_config": {
      "zone":  "hk-02",
      "image_id_filter":  {
        "os_type": "Linux",
        "image_type": "Base",
        "most_recent": true
      },
      "name_prefix":  "Test-migrate",
      "password": "ucloud_2022",
      "charge_type":  "Dynamic",
      "cpu":  1,
      "memory": 1024,
      "tag":  "migrate",
      "minimal_cpu_platform": "Amd/Auto",
      "machine_type":  "OM",
      "disks": [
        {
          "is_boot": "True",
          "size": 20,
          "type": "CLOUD_RSSD"
        }
      ]
    },
    "cube_config": {
      "cube_id_filter": {
        "zone": "hk-02"
      }
    }
  }
}
```

## Config Doc
You may refer to the [API Docs](https://docs.ucloud.cn/api):
- [CreateUHostInstance](https://docs.ucloud.cn/api/uhost-api/create_uhost_instance)
- [DescribeImage](https://docs.ucloud.cn/api/uhost-api/describe_image)
- [ListCubePod](https://docs.ucloud.cn/api/cube-api/list_cube_pod)

### Argument Reference

* `public_key` - (Required) This is the UCloud public key. You may refer to [get public key from console](https://console.ucloud.cn/uapi/apikey). It must be provided, but
  it can also be sourced from the `UCLOUD_PUBLIC_KEY` environment variable.

* `private_key` - (Required) This is the UCloud private key. You may refer to [get private key from console](https://console.ucloud.cn/uapi/apikey). It must be provided, but
  it can also be sourced from the `UCLOUD_PRIVATE_KEY` environment variable.

* `region` - (Required) This is the UCloud region. It must be provided, but
  it can also be sourced from the `UCLOUD_REGION` environment variables.

* `project_id` - (Required) This is the UCloud project id. It must be provided, but
  it can also be sourced from the `UCLOUD_PROJECT_ID` environment variables. 

* `migrate_eip`- (Required when use cmd `migrate eip`) See [migrate_eip](#migrate_eip) below for details on attributes.

* `migrate_ulb`- (Required when use cmd `migrate ulb`) See [migrate_ulb](#migrate_ulb) below for details on attributes.

* `migrate_private_ip`- (Required when use cmd `migrate private-ip`) See [migrate_private_ip](#migrate_private_ip) below for details on attributes.

* `log` - (Optional) See [log](#log) below for details on attributes.(Default: Stdout)

#### log

* `is_stdout` - (Optional) This bool value to control is or not print log to Stdout.(Default: `false`)
* `dir` - (Optional) Set the dir of log file located for not Stdout. (Default: `./build`)
* `name` - (Optional) Set the log file name for not Stdout. (Default: `migrate`)
* `level` - (Optional) Set the log level, Possible values are: `DEBUG`, `INFO`, `WARN`, `ERROR`. (Default: `DEBUG`)

#### migrate_eip

* `uhost_config` - (Required) See [uhost_config](#uhost_config) below for details on attributes.
* `cube_config` - (Required) See [cube_config](#cube_config) below for details on attributes.
* `service_validation` - (Optional) See [service_validation](#service_validation) below for details on attributes.

#### migrate_ulb

* `ulb_id` - (Required) The id of ulb instance.
* `uhost_config` - (Required) See [uhost_config](#uhost_config) below for details on attributes.
* `service_validation` - (Optional) See [service_validation](#service_validation) below for details on attributes.

#### migrate_private_ip

* `uhost_config` - (Required) See [uhost_config](#uhost_config) below for details on attributes.
* `cube_config` - (Required) See [cube_config](#cube_config) below for details on attributes.

#### cube_config

* `cube_id_list` - (Optional, Array) The cube id list.
* `cube_id_filter` - (Optional) See [cube_id_filter](#cube_id_filter) below for details on attributes.

#### cube_id_filter

* `zone` - (Optional) Availability zone where Cube instance is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `image_id` - (Required) The ID for the image to use for the instance.
* `vpc_id` - (Optional) The ID of VPC linked to the cube instance.
* `subnet_id` - (Optional) The ID of subnet linked to the cube instance.
* `group` - (Optional) The group of cube instance.
* `deployment_id` - (Optional) The deployment ID of the cube instance.
* `name_regex` - (Optional) The regex string to filter resulting cube by name.

#### uhost_config

* `zone` - (Required) Availability zone where UHost instance is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `disks` - (Required, Array) See [disks](#disks) below for details on attributes.
* `password` - (Required) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will auto-generate a password.
* `image_id` - (Optional) The ID for the image to use for the UHost instance.(must set one of `image_id` and `image_id_fiter`)
* `image_id_filter` - (Optional) See [image_id_filter](#image_id_filter) below for details on attributes. 
* `name` - (Optional) The name of UHost instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'.
* `name_prefix` - (Optional) The name prefix of UHost instance. If not specified one of `name` and `name_prefix`, this tool will auto-generate a name beginning with `uhost-instance`.
* `charge_type` - (Optional) The charge type of UHost instance, possible values are: `Year`, `Month` and `Dynamic` as pay by hour (specific permission required). (Default: `Month`).
* `duration` - (Optional) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `Dynamic` (pay by hour).
* `cpu` - (Optional) The CPU of UHost Instance.(1-64, Default: `1`)
* `memory` - (Optional) The Memory of UHost Instance.(1024-262144, Default: `1024`)
* `tag` - (Optional) A tag assigned to instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `minimal_cpu_platform` - (Optional) Specifies a minimum CPU platform for the VM instance. (Default: `Amd/Auto`). You may refer to [min_cpu_platform](https://docs.ucloud.cn/uhost/introduction/uhost/type_new)
* `machine_type` - (Optional) The Machine type of UHost instance.(Default: `OM`)
* `security_group_id` - (Optional) The ID of the associated firewall.
* `vpc_id`- (Optional) The ID of VPC linked to the UHost instance.(can not use for cmd `migrate private-ip`)
* `subenet_id`- (Optional) The ID of subnet linked to the UHost instance.(can not use for cmd `migrate private-ip`)

#### disks

* `is_boot` - (Required) The string value to set the disk is or not boot system disk. `True` means boot system disk, `False` means data disk.
* `size` - (Required) The size of the data disk, range 20-8000, measured in GB (GigaByte).
* `type` - (Required) The type of disk, you may refer to [disk_type](https://docs.ucloud.cn/api/uhost-api/disk_type).

#### image_id_filter

* `os_type` - (Optional) The type of OS. Possible values are: `Linux` and `Windows`, all the OS types will be retrieved by default.
* `image_type` - (Optional) The type of image. Possible values are: `Base` as standard image, `Business` as owned by marketplace, and `Custom` as custom-image, all the image types will be retrieved by default.
* `most_recent` - (Optional) If more than one result is returned, use the most recent image.
* `name_regex` - (Optional) A regex string to filter resulting images by name. (Such as: `^CentOS 8.[2-3] 64` means CentOS 8.2 of 64-bit operating system or CentOS 8.3 of 64-bit operating system).

#### service_validation

* `Port` - (Required for cmd `migrate eip`) This field only for cmd `migrate eip`.
* `WaitServiceReadyTimeout` - (Required) The time limit for wait service ready.(Default: `120` time second)

## Warning

The migrate tool not support distributed consistency service. Therefore, we need to ensure that the CLI can be executed completely without interruption. 
- [`migrate eip`] If the interruption leads to the unbinding of the EIP, you can query the log and manually bind it, which will cause the service provided by the one IP to be unavailable.
There into, you can use [UCloud CLI](https://docs.ucloud.cn/cli/README) cmd to band eip.
```shell
ucloud eip bind --eip-id "xxx" --resource-type "cube" --resource-id "xxx"
```
- [`migrate ulb`] If the interruption leads to the deleting ulb backend about cube, you can query the log and manually delete it.
    There into, you can use [UCloud CLI](https://docs.ucloud.cn/cli/README) cmd to band eip.
```shell
ucloud ulb vserver backend delete --backend-id "xxx" --ulb-id "xxx"
```
- [`migrate private-ip`] If the interruption leads to the cube have been deleted but the UHost have not been created with the private ip, you can query the log and manually create UHost with private ip.
There into, you can use [UCloud CLI](https://docs.ucloud.cn/cli/README) cmd to create UHost and you can refer the [UAPI](https://console.ucloud.cn/uapi/detail?id=CreateUHostInstance).
```shell
ucloud api \
  --Action CreateUHostInstance \
  --Region hk \
  --Zone hk-02 \
  --ProjectId org-xxx \
  --VPCId xxx \
  --SubnetId xxx \
  --ImageId uimage-xxx \
  --LoginMode Password \
  --Password dWNsb3VkXzIwMjI= \
  --Name Test-migrate \
  --Tag migrate \
  --CPU 1 \
  --Memory 1024 \
  --MachineType OM \
  --MinimalCpuPlatform Amd/Auto \
  --Disks.0.IsBoot True \
  --Disks.0.Type CLOUD_RSSD \
  --Disks.0.Size 20 \
  --PrivateIp.0 xxx
```
