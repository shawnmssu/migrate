# Migrate

Migrate CLI provides a unified command line interface for migrating Cube instances to UHost instances at the present. 
The Process about:
- Create UHost List (Default: Shared Outstanding UHost)
- Waiting for UHost Running
- Unbind one of the queried Cube with EIP and then bind EIP to UHost
- Repeat the previous step
- [Option]Running tcp validation about UHost service.
- [Option]RollBack the EIP to Cube When migrate got error.

## Installation

- Go 1.16
- You have installed git and golang on your platform, you can fetch the source code from GitHub and compile it by yourself.

```bash
git clone https://github.com/shawnmssu/migrate.git
cd migrate
make linux # search one system
./bin/migrate --help
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
- for example:
```json
{
  "public_key": "xxx",
  "private_key": "xxx",
  "project_id": "org-xxx",
  "region": "hk",
  "migrate": {
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

## Warning

The migrate tool not support distributed consistency service. Therefore, we need to ensure that the CLI can be executed completely without interruption. 
If the interruption leads to the unbinding of the EIP, you can query the log and manually bind it, which will cause the service provided by the one IP to be unavailable.
There into, you can use [UCloud CLI](https://docs.ucloud.cn/cli/README) cmd to band eip.
```shell
ucloud eip bind --eip-id "xxx" --resource-type "cube" --resource-id "xxx"
```
