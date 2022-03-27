# -*- mode: Python -*-

load('ext://restart_process', 'docker_build_with_restart')

docker_build(
  ref='algo_testnet',
  context='./algorand',
  dockerfile='deployments/Dockerfile.algo',
)

docker_build(
  ref='algo_indexer',
  context='./algorand',
  dockerfile='deployments/Dockerfile.algo_indexer',
)

k8s_yaml('deployments/algo_testnet_k8s.yaml')

k8s_resource(
  'algorand',
  port_forwards = [
      port_forward(4001, name = "Algorand RPC", host = 'localhost'),
      port_forward(4002, name = "Algorand KMD RPC", host = 'localhost'),
      port_forward(4003, name = "Algorand Indexer", host = 'localhost'),
  ],
)

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/http-server-go ./http_server'

local_resource(
  'http-server-go-compile',
  compile_cmd,
  deps=['./https_server/main.go', './http_server/ops'],
)

docker_build_with_restart(
  ref='http-server-image',
  context='.',
  entrypoint=['/app/build/http-server-go', '--algod_host=http://algorand:4001'],
  dockerfile='deployments/Dockerfile.http_server',
  only=[
    './algorand',
    './build',
    './http_server/priv',
    './http_server/static',
    './http_server/templates',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./http_server/priv', '/app/priv'),
    sync('./http_server/static', '/app/static'),
    sync('./http_server/templates', '/app/templates'),
  ],
)

k8s_yaml('deployments/http_server_k8s.yaml')
k8s_resource(
  'http-server',
  port_forwards=1234,
  resource_deps=['http-server-go-compile', 'algorand'],
)

