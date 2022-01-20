# Migrate

Migrate CLI provides a unified command line interface for migrating Cube instances to UHost instances at the present. 

The Process about:
- `migrate eip`
  - Create UHost List (Default: Shared Outstanding UHost)
  - Waiting for UHost Running
  - Unbind one of the queried Cube with EIP and then bind EIP to UHost
  - Repeat the previous step
  - [Option]Running tcp validation about UHost service.
  - [Option]RollBack the EIP to Cube When migrate got error.
- `migrate ulb`
  - Create UHost List (Default: Shared Outstanding UHost)
  - Waiting for UHost Running
  - Create one ULB VServer Backend used UHost and then delete one ulb backend about Cube.(migrate ulb policy not supported yet)
  - Repeat the previous step
  - [Option]Running ulb backend health check
  - [Option]RollBack the ulb backend about Cube When migrate got error.
- `migrate private-ip`
  - Try to Create temporary one UHost to validate UHost Config, waiting for UHost running and then delete it.
  - Delete one Cube for freed the private ip
  - Create UHost By the private ip and UHost Config (Default: Shared Outstanding UHost)
  - Waiting for UHost Running
  - Repeat the previous step
  - [Option] use the `dry-run` flag to validate UHost Config and Cube config 
    - if set true which means that only try to create the UHost with config, waiting for UHost running and then delete it.
    - for example:`migrate private-ip --conf xxx ----dry-run` 

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

## Config

- You may refer to the [API Docs](https://docs.ucloud.cn/api):
  - [CreateUHostInstance](https://docs.ucloud.cn/api/uhost-api/create_uhost_instance)
  - [DescribeImage](https://docs.ucloud.cn/api/uhost-api/describe_image)
  - [ListCubePod](https://docs.ucloud.cn/api/cube-api/list_cube_pod)
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
