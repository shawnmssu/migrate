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
  - [Option] use the `dry-run` flag to validate UHost Config and Cube config.
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

* `migrate_private_ip`- (Required when use cmd `migrate private-ip`) See [migrate_private_ip](#migrate_private_ip) below for details on attributes.

#### migrate_private_ip

* `uhost_config` - (Required) See [uhost_config](#uhost_config) below for details on attributes.
* `cube_config` - (Required) See [cube_config](#cube_config) below for details on attributes.

#### cube_config

* `cube_id_list` - (Optional, Array)
* `cube_id_filter` - (Optional) See [cube_id_filter](#cube_id_filter) below for details on attributes.

#### cube_id_filter

* `zone`
* `vpc_id`
* `subnet_id`
* `group`
* `deployment_id`
* `name_regex`

#### uhost_config

* `zone` - (Required)
* `disks` - (Required, Array) See [disks](#disks) below for details on attributes.
* `password` - (Required)
* `image_id` - (Optional)
* `image_id_filter` - (Optional) See [image_id_filter](#image_id_filter) below for details on attributes.
* `name` - (Optional)
* `name_prefix` - (Optional)
* `charge_type` - (Optional)
* `duration` - (Optional)
* `cpu` - (Optional)
* `memory` - (Optional)
* `tag` - (Optional)
* `minimal_cpu_platform` - (Optional)
* `machine_type` - (Optional)
* `security_group_id` - (Optional)

#### disks

* `is_boot` - (Required)
* `size` - (Required)
* `type` - (Required)

#### image_id_filter

* `os_type` - (Optional)
* `image_type` - (Optional)
* `most_recent` - (Optional)
* `name_regex` - (Optional)

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
