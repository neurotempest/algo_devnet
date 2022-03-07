# -*- mode: Python -*-

docker_build(
  ref='algo_testnet_docker_image',
  context='.',
  dockerfile='deployments/Dockerfile',
)

k8s_yaml('deployments/algo_testnet_k8s.yaml')

k8s_resource(
  'algorand-testnet',
  port_forwards = [
      port_forward(8080, name = "Algorand RPC", host = 'localhost'),
  ],
)
