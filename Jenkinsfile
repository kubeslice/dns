@Library('jenkins-library@opensource') _
dockerImagePipeline(
  script: this,
  service: 'dns',
  dockerfile: 'Dockerfile',
  buildContext: '.',
  buildArguments: [PLATFORM:"amd64"]
  
)
